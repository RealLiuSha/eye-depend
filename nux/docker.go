package nux

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"errors"
	"context"
	"strconv"
	"strings"
	"encoding/json"

	"github.com/fsouza/go-dockerclient"

	"github.com/itchenyi/str"
	"git.wolaidai.com/DevOps/eye-depend/sys"
	"git.wolaidai.com/DevOps/eye-depend/bytefmt"
)

type DockerStat struct {
	Id          string
	Name        string
	Container   string
	CpuPercent  float64
	MemUsage    uint64
	MemTotal    uint64
	MemPercent  float64
	NetInput    uint64
	NetOutput   uint64
	DiskRead    uint64
	DiskWrite   uint64
}

type DockerInfo struct {
	App         string
	Task        string
	Ip          string
}

func CurrentDockerStats() ([]DockerStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(15))
	defer cancel()

	cmdJsonFormat := `--format '{"container":"{{ .Container }}",` +
		`"id":"{{.ID}}","name":"{{.Name}}","cpu.percent":"{{.CPUPerc}}",` +
		`"mem.usage":"{{.MemUsage}}","mem.percent":"{{.MemPerc}}",` +
		`"net.io":"{{.NetIO}}","block.io":"{{.BlockIO}}"}'`

	cmd := exec.CommandContext(
		ctx, "bash", "-c",
		str.JoinStrings("docker stats --no-stream ", cmdJsonFormat))

	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errors.New("Execution timed out")
	}

	if err != nil {
		return nil, fmt.Errorf("Non-zero exit code: %s", err)
	}

	dockerStatList := make([]DockerStat, 0)
	for _, stat := range strings.Split(string(out), "\n") {
		if stat == "" { continue }

		dockerStat, err := parseDockerStat([]byte(stat))
		if err != nil {
			fmt.Println(err)
			continue
		}

		dockerStatList = append(dockerStatList, dockerStat)
	}

	if len(dockerStatList) == 0 {
		return nil, errors.New("Get Docker stats failed.")
	}

	return dockerStatList, nil
}

func valueReplacer(value string) string {
	replacer := strings.NewReplacer(" ", "", "i", "")
	return replacer.Replace(value)
}

func valueSplit(value string) (string, string, error) {
	if values := strings.Split(value, "/"); len(values) == 2 {
		return valueReplacer(values[0]), valueReplacer(values[1]), nil
	}

	return "", "", errors.New("invalid value, cannot be split.")
}

func parseValue2Percent(value string) (float64, error) {
	number, err := strconv.ParseFloat(strings.Trim(value, "%"), 64)
	if err != nil {
		return 0.0, fmt.Errorf("parseCpuPercent Error: %s", err)
	}

	return number, nil
}

func parseValue2Byte(value string) (uint64, uint64, error) {
	var leftVal, rightVal uint64
	strLeftVal, strRightVal, err := valueSplit(value)
	if err != nil {
		return 0, 0, err
	}

	leftVal, err = bytefmt.ToBytes(strLeftVal)
	if err != nil {
		return 0, 0, err
	}

	rightVal, err = bytefmt.ToBytes(strRightVal)
	if err != nil {
		return 0, 0, err
	}

	return leftVal, rightVal, nil
}

func parseDockerStat(statByte []byte) (DockerStat, error) {
	var stat map[string]string
	var dockerStat DockerStat

	if err := json.Unmarshal(statByte, &stat); err != nil {
		return dockerStat, err
	}

	if id, ok := stat["id"]; ok {
		dockerStat.Id = id
	}

	if name, ok := stat["name"]; ok {
		dockerStat.Name = name
	}

	if container, ok := stat["container"]; ok {
		dockerStat.Container = container
	}

	if cpuPercent, ok := stat["cpu.percent"]; ok {
		value, err := parseValue2Percent(cpuPercent)
		if err != nil {
			return dockerStat, err
		}

		dockerStat.CpuPercent = value
	}

	if memPercent, ok := stat["mem.percent"]; ok {
		value, err := parseValue2Percent(memPercent)
		if err != nil {
			return dockerStat, err
		}

		dockerStat.MemPercent = value
	}

	if memUsage, ok := stat["mem.usage"]; ok {
		usageVal, totalVal, err := parseValue2Byte(memUsage)
		if err != nil {
			return dockerStat, err
		}

		dockerStat.MemUsage = usageVal
		dockerStat.MemTotal = totalVal
	}

	if netIo, ok := stat["net.io"]; ok {
		netInput, netOutput, err := parseValue2Byte(netIo)
		if err != nil {
			return dockerStat, err
		}

		dockerStat.NetInput = netInput
		dockerStat.NetOutput = netOutput
	}

	if blockIo, ok := stat["block.io"]; ok {
		diskRead, diskWrite, err := parseValue2Byte(blockIo)
		if err != nil {
			return dockerStat, err
		}

		dockerStat.DiskRead = diskRead
		dockerStat.DiskWrite = diskWrite
	}

	return dockerStat, nil
}

func CurrentDockerInfo(id string, client *docker.Client) (DockerInfo, error) {
	var dockerInfo DockerInfo

	container, err := client.InspectContainer(id)
	if err != nil {
		return dockerInfo, err
	}

	getContainerIP := func () string {
		for key := range container.NetworkSettings.Networks {
			if obj, isExist := container.NetworkSettings.Networks[key]; isExist {
				if obj.IPAddress != "" {
					return obj.IPAddress
				}

				continue
			}
		}

		ifconfig := fmt.Sprintf("docker exec %s ip addr show ", id) +
			"eth0|awk '/inet /{print $2}'|cut -d/ -f1"

		address, err := sys.CmdOut("sh", "-c", ifconfig)
		if err != nil {
			return "255.255.255.0"
		}

		return strings.TrimSuffix(address, "\n")
	}

	dockerInfo.Ip = getContainerIP()
	dockerInfo.App = container.Config.Labels["SRV_NAME"]
	dockerInfo.Task = container.Config.Labels["MESOS_TASK_ID"]

	return dockerInfo, nil
}