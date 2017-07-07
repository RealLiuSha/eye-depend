package dkstat

import (
	"io"
	"github.com/fsouza/go-dockerclient"
)

type Writer struct {
	writer   io.Writer
}

type Metrics struct {
	CPU	struct {
		UsageInUsermode			uint64
		UsageInKernelmode		uint64
		TotalUsage				uint64
	}

	Mem struct {
		Limit					uint64
		MaxUsage				uint64
		Usage					uint64
		TotalActiveAnon			uint64
		TotalActiveFile			uint64
		TotalCache				uint64
		TotalInactiveAnon		uint64
		TotalInactiveFile		uint64
		TotalMappedFile			uint64
		TotalPgfault			uint64
		TotalPgpgin				uint64
		TotalPgpgout			uint64
		TotalRss				uint64
		TotalRssHuge			uint64
		TotalUnevictable		uint64
		TotalWriteback			uint64
	}

	Net map[string]docker.NetworkStats
	App string
	Task string
	ID   string
	IP   string
}

func NewWriter(writer io.Writer) Writer {
	return Writer{
		writer: writer,
	}
}

func (writer Writer) GetMetrics(stats Stats) Metrics {
	var metrics Metrics

	metrics.CPU.UsageInUsermode       = stats.Stats.CPUStats.CPUUsage.UsageInUsermode
	metrics.CPU.UsageInKernelmode     = stats.Stats.CPUStats.CPUUsage.UsageInKernelmode
	metrics.CPU.TotalUsage            = stats.Stats.CPUStats.CPUUsage.TotalUsage
	metrics.Mem.Limit                 = stats.Stats.MemoryStats.Limit
	metrics.Mem.MaxUsage              = stats.Stats.MemoryStats.MaxUsage
	metrics.Mem.Usage                 = stats.Stats.MemoryStats.Usage

	metrics.Mem.Usage                 = stats.Stats.MemoryStats.Stats.TotalActiveAnon
	metrics.Mem.TotalActiveFile       = stats.Stats.MemoryStats.Stats.TotalActiveFile
	metrics.Mem.TotalCache            = stats.Stats.MemoryStats.Stats.TotalCache
	metrics.Mem.TotalInactiveAnon     = stats.Stats.MemoryStats.Stats.TotalInactiveAnon
	metrics.Mem.TotalInactiveFile     = stats.Stats.MemoryStats.Stats.TotalInactiveFile
	metrics.Mem.TotalMappedFile       = stats.Stats.MemoryStats.Stats.TotalMappedFile
	metrics.Mem.TotalPgfault          = stats.Stats.MemoryStats.Stats.TotalPgfault
	metrics.Mem.TotalPgpgin           = stats.Stats.MemoryStats.Stats.TotalPgpgin
	metrics.Mem.TotalPgpgout          = stats.Stats.MemoryStats.Stats.TotalPgpgout
	metrics.Mem.TotalRss              = stats.Stats.MemoryStats.Stats.TotalRss
	metrics.Mem.TotalRssHuge          = stats.Stats.MemoryStats.Stats.TotalRssHuge
	metrics.Mem.TotalUnevictable      = stats.Stats.MemoryStats.Stats.TotalUnevictable
	metrics.Mem.TotalWriteback        = stats.Stats.MemoryStats.Stats.TotalWriteback
	metrics.Net						  = stats.Stats.Networks
	metrics.Task					  = stats.Task
	metrics.App   					  = stats.App
	metrics.ID    					  = stats.ID
	metrics.IP						  = stats.IP

	return metrics
}
