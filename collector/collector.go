package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "dockhand"

// HostStats holds the JSON response from /api/dashboard/stats for one host.
type HostStats struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`

	Online bool `json:"online"`

	Containers struct {
		Total          int `json:"total"`
		Running        int `json:"running"`
		Stopped        int `json:"stopped"`
		Paused         int `json:"paused"`
		Restarting     int `json:"restarting"`
		Unhealthy      int `json:"unhealthy"`
		PendingUpdates int `json:"pendingUpdates"`
	} `json:"containers"`

	Images struct {
		Total     int   `json:"total"`
		TotalSize int64 `json:"totalSize"`
	} `json:"images"`

	Volumes struct {
		Total     int   `json:"total"`
		TotalSize int64 `json:"totalSize"`
	} `json:"volumes"`

	ContainersSize int64 `json:"containersSize"`
	BuildCacheSize int64 `json:"buildCacheSize"`

	Networks struct {
		Total int `json:"total"`
	} `json:"networks"`

	Stacks struct {
		Total   int `json:"total"`
		Running int `json:"running"`
		Partial int `json:"partial"`
		Stopped int `json:"stopped"`
	} `json:"stacks"`

	Metrics *struct {
		CPUPercent    float64 `json:"cpuPercent"`
		MemoryPercent float64 `json:"memoryPercent"`
		MemoryUsed    int64   `json:"memoryUsed"`
		MemoryTotal   int64   `json:"memoryTotal"`
	} `json:"metrics"`

	Events struct {
		Total int `json:"total"`
		Today int `json:"today"`
	} `json:"events"`
}

// DockhandCollector implements prometheus.Collector.
type DockhandCollector struct {
	url   string
	token string

	// host status
	hostOnline *prometheus.Desc

	// containers
	containersTotal          *prometheus.Desc
	containersRunning        *prometheus.Desc
	containersStopped        *prometheus.Desc
	containersPaused         *prometheus.Desc
	containersRestarting     *prometheus.Desc
	containersUnhealthy      *prometheus.Desc
	containersPendingUpdates *prometheus.Desc

	// images
	imagesTotal     *prometheus.Desc
	imagesSizeBytes *prometheus.Desc

	// volumes
	volumesTotal     *prometheus.Desc
	volumesSizeBytes *prometheus.Desc

	// sizes
	containersSizeBytes *prometheus.Desc
	buildCacheSizeBytes *prometheus.Desc

	// networks
	networksTotal *prometheus.Desc

	// stacks
	stacksTotal   *prometheus.Desc
	stacksRunning *prometheus.Desc
	stacksPartial *prometheus.Desc
	stacksStopped *prometheus.Desc

	// metrics
	cpuPercent    *prometheus.Desc
	memoryPercent *prometheus.Desc
	memoryUsed    *prometheus.Desc
	memoryTotal   *prometheus.Desc

	// events
	eventsTotal *prometheus.Desc
	eventsToday *prometheus.Desc
}

var hostLabels = []string{"id", "name", "host"}

func newDesc(subsystem, name, help string, variableLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help,
		variableLabels,
		nil,
	)
}

// New returns a DockhandCollector that scrapes the given Dockhand instance.
func New(url, token string) *DockhandCollector {
	return &DockhandCollector{
		url:   url,
		token: token,

		hostOnline: newDesc("host", "online", "1 if the Docker host is reachable, 0 otherwise.", hostLabels),

		containersTotal:          newDesc("containers", "total", "Total number of containers.", hostLabels),
		containersRunning:        newDesc("containers", "running", "Number of running containers.", hostLabels),
		containersStopped:        newDesc("containers", "stopped", "Number of stopped containers.", hostLabels),
		containersPaused:         newDesc("containers", "paused", "Number of paused containers.", hostLabels),
		containersRestarting:     newDesc("containers", "restarting", "Number of restarting containers.", hostLabels),
		containersUnhealthy:      newDesc("containers", "unhealthy", "Number of unhealthy containers.", hostLabels),
		containersPendingUpdates: newDesc("containers", "pending_updates", "Number of containers with a pending image update.", hostLabels),

		imagesTotal:     newDesc("images", "total", "Total number of images.", hostLabels),
		imagesSizeBytes: newDesc("images", "size_bytes", "Total disk size of all images in bytes.", hostLabels),

		volumesTotal:     newDesc("volumes", "total", "Total number of volumes.", hostLabels),
		volumesSizeBytes: newDesc("volumes", "size_bytes", "Total disk size of all volumes in bytes.", hostLabels),

		containersSizeBytes: newDesc("containers", "size_bytes", "Total disk size used by container writable layers in bytes.", hostLabels),
		buildCacheSizeBytes: newDesc("build_cache", "size_bytes", "Total disk size used by the build cache in bytes.", hostLabels),

		networksTotal: newDesc("networks", "total", "Total number of networks.", hostLabels),

		stacksTotal:   newDesc("stacks", "total", "Total number of stacks.", hostLabels),
		stacksRunning: newDesc("stacks", "running", "Number of running stacks.", hostLabels),
		stacksPartial: newDesc("stacks", "partial", "Number of partially running stacks.", hostLabels),
		stacksStopped: newDesc("stacks", "stopped", "Number of stopped stacks.", hostLabels),

		cpuPercent:    newDesc("host", "cpu_percent", "CPU usage percentage of the Docker host.", hostLabels),
		memoryPercent: newDesc("host", "memory_percent", "Memory usage percentage of the Docker host.", hostLabels),
		memoryUsed:    newDesc("host", "memory_used_bytes", "Memory used on the Docker host in bytes.", hostLabels),
		memoryTotal:   newDesc("host", "memory_total_bytes", "Total memory of the Docker host in bytes.", hostLabels),

		eventsTotal: newDesc("events", "total", "Total number of recorded Docker events.", hostLabels),
		eventsToday: newDesc("events", "today", "Number of Docker events recorded today.", hostLabels),
	}
}

// Describe sends all metric descriptors to the channel.
func (c *DockhandCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hostOnline
	ch <- c.containersTotal
	ch <- c.containersRunning
	ch <- c.containersStopped
	ch <- c.containersPaused
	ch <- c.containersRestarting
	ch <- c.containersUnhealthy
	ch <- c.containersPendingUpdates
	ch <- c.imagesTotal
	ch <- c.imagesSizeBytes
	ch <- c.volumesTotal
	ch <- c.volumesSizeBytes
	ch <- c.containersSizeBytes
	ch <- c.buildCacheSizeBytes
	ch <- c.networksTotal
	ch <- c.stacksTotal
	ch <- c.stacksRunning
	ch <- c.stacksPartial
	ch <- c.stacksStopped
	ch <- c.cpuPercent
	ch <- c.memoryPercent
	ch <- c.memoryUsed
	ch <- c.memoryTotal
	ch <- c.eventsTotal
	ch <- c.eventsToday
}

// Collect fetches metrics from the Dockhand API and sends them to the channel.
func (c *DockhandCollector) Collect(ch chan<- prometheus.Metric) {
	hosts, err := c.fetchStats()
	if err != nil {
		log.Printf("error fetching dockhand stats: %v", err)
		return
	}

	for _, h := range hosts {
		id := fmt.Sprintf("%d", h.ID)
		name := h.Name
		host := h.Host
		if host == "" {
			host = strings.ToLower(name)
		}
		labels := []string{id, name, host}

		online := 0.0
		if h.Online {
			online = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.hostOnline, prometheus.GaugeValue, online, labels...)

		ch <- prometheus.MustNewConstMetric(c.containersTotal, prometheus.GaugeValue, float64(h.Containers.Total), labels...)
		ch <- prometheus.MustNewConstMetric(c.containersRunning, prometheus.GaugeValue, float64(h.Containers.Running), labels...)
		ch <- prometheus.MustNewConstMetric(c.containersStopped, prometheus.GaugeValue, float64(h.Containers.Stopped), labels...)
		ch <- prometheus.MustNewConstMetric(c.containersPaused, prometheus.GaugeValue, float64(h.Containers.Paused), labels...)
		ch <- prometheus.MustNewConstMetric(c.containersRestarting, prometheus.GaugeValue, float64(h.Containers.Restarting), labels...)
		ch <- prometheus.MustNewConstMetric(c.containersUnhealthy, prometheus.GaugeValue, float64(h.Containers.Unhealthy), labels...)
		ch <- prometheus.MustNewConstMetric(c.containersPendingUpdates, prometheus.GaugeValue, float64(h.Containers.PendingUpdates), labels...)

		ch <- prometheus.MustNewConstMetric(c.imagesTotal, prometheus.GaugeValue, float64(h.Images.Total), labels...)
		ch <- prometheus.MustNewConstMetric(c.imagesSizeBytes, prometheus.GaugeValue, float64(h.Images.TotalSize), labels...)

		ch <- prometheus.MustNewConstMetric(c.volumesTotal, prometheus.GaugeValue, float64(h.Volumes.Total), labels...)
		ch <- prometheus.MustNewConstMetric(c.volumesSizeBytes, prometheus.GaugeValue, float64(h.Volumes.TotalSize), labels...)

		ch <- prometheus.MustNewConstMetric(c.containersSizeBytes, prometheus.GaugeValue, float64(h.ContainersSize), labels...)
		ch <- prometheus.MustNewConstMetric(c.buildCacheSizeBytes, prometheus.GaugeValue, float64(h.BuildCacheSize), labels...)

		ch <- prometheus.MustNewConstMetric(c.networksTotal, prometheus.GaugeValue, float64(h.Networks.Total), labels...)

		ch <- prometheus.MustNewConstMetric(c.stacksTotal, prometheus.GaugeValue, float64(h.Stacks.Total), labels...)
		ch <- prometheus.MustNewConstMetric(c.stacksRunning, prometheus.GaugeValue, float64(h.Stacks.Running), labels...)
		ch <- prometheus.MustNewConstMetric(c.stacksPartial, prometheus.GaugeValue, float64(h.Stacks.Partial), labels...)
		ch <- prometheus.MustNewConstMetric(c.stacksStopped, prometheus.GaugeValue, float64(h.Stacks.Stopped), labels...)

		if h.Metrics != nil {
			ch <- prometheus.MustNewConstMetric(c.cpuPercent, prometheus.GaugeValue, h.Metrics.CPUPercent, labels...)
			ch <- prometheus.MustNewConstMetric(c.memoryPercent, prometheus.GaugeValue, h.Metrics.MemoryPercent, labels...)
			ch <- prometheus.MustNewConstMetric(c.memoryUsed, prometheus.GaugeValue, float64(h.Metrics.MemoryUsed), labels...)
			ch <- prometheus.MustNewConstMetric(c.memoryTotal, prometheus.GaugeValue, float64(h.Metrics.MemoryTotal), labels...)
		}

		ch <- prometheus.MustNewConstMetric(c.eventsTotal, prometheus.CounterValue, float64(h.Events.Total), labels...)
		ch <- prometheus.MustNewConstMetric(c.eventsToday, prometheus.GaugeValue, float64(h.Events.Today), labels...)
	}
}

func (c *DockhandCollector) fetchStats() ([]HostStats, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, c.url+"/api/dashboard/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var hosts []HostStats
	if err := json.Unmarshal(body, &hosts); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return hosts, nil
}
