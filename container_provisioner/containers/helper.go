package containers

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/database"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	containerImage = "ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest"
)

// initializeDockerClient initialize a new docker api client
func initializeDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	return cli
}

// PullImage pulls the given image from a registry
func PullImage(image string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	reader, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	utils.ErrorHandler(err)
	defer reader.Close()

	// Print the progress of the image pull
	_, err = io.Copy(os.Stdout, reader)
	utils.ErrorHandler(err)
}

// RemoveContainer removes the container with the given ID
func RemoveContainer(containerID string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Remove the container
	err = cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	utils.ErrorHandler(err)
}

// ContainerConfigGenerator generates the container config depending on the scrape target
func ContainerConfigGenerator(
	scrapeTarget string,
	scrapeTargetName string,
	scrapeURL string, uploadIdentifier string,
	proxyAddress string, proxyLocation string, proxyPort string) *container.Config {

	var scrapeContainerURL string
	var targetName string

	switch scrapeTarget {
	case "HOTEL":
		scrapeContainerURL = fmt.Sprintf("HOTEL_URL=%s", scrapeURL)
		targetName = fmt.Sprintf("HOTEL_NAME=%s", scrapeTargetName)
	case "RESTO":
		scrapeContainerURL = fmt.Sprintf("RESTO_URL=%s", scrapeURL)
		targetName = fmt.Sprintf("RESTO_NAME=%s", scrapeTargetName)
	case "AIRLINE":
		scrapeContainerURL = fmt.Sprintf("AIRLINE_URL=%s", scrapeURL)
		targetName = fmt.Sprintf("AIRLINE_NAME=%s", scrapeTargetName)
	}

	scrapeMode := fmt.Sprintf("SCRAPE_MODE=%s", scrapeTarget)
	proxySettings := fmt.Sprintf("PROXY_ADDRESS=socks5://%s:%s", proxyAddress, proxyPort)

	return &container.Config{
		Image: containerImage,
		Labels: map[string]string{
			"TaskOwner":  uploadIdentifier,
			"Target":     scrapeTargetName,
			"vpn.region": proxyLocation,
		},
		// Env vars required by the js scraper containers
		Env: []string{
			"CONCURRENCY=2",
			"IS_PROVISIONER=true",
			scrapeMode,
			scrapeContainerURL,
			targetName,
			proxySettings,
		},
		Tty: true,
	}
}

// CreateContainer creates a container then returns the container ID
func CreateContainer(containerConfig *container.Config) string {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Create the container. Container.ID contains the ID of the container
	Container, err := cli.ContainerCreate(context.Background(),
		containerConfig,
		&container.HostConfig{
			AutoRemove: false, // Cant set to true otherwise the container got deleted before copying the file
		},
		nil, // NetworkConfig
		nil, // Platform
		"",  // Container name
	)

	utils.ErrorHandler(err)

	return Container.ID[:12]
}

// TailLog tails the log of the container with the given ID
func TailLog(containerID string) io.Reader {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Print the logs of the container
	out, err := cli.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		Details:    true,
		ShowStderr: false,
		Timestamps: false,
		Follow:     true})
	utils.ErrorHandler(err)

	// // Docker log uses multiplexed streams to send stdout and stderr in the connection. This function separates them
	// _, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	// utils.ErrorHandler(err)

	return out
}

// Container information
type Container struct {
	ContainerID  *string
	TaskOwner    *string
	TargetName   *string
	URL          *string
	IPAddress    *string
	VPNRegion    *string
	VPNSOCKSPort *string
	VPNHTTPPort  *string
}

// ListContainersByType lists all containers of the given type.
// Available container types:
//   - "scraper": Lists all scraper containers.
//   - "proxy": Lists all proxy containers.
//
// Example:
//
//	scraperContainers := ListContainersByType("scraper")
//	proxyContainers := ListContainersByType("proxy")
func ListContainersByType(containerType string) []Container {

	// Initialize a new docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// List all containers
	containersInfo, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	utils.ErrorHandler(err)

	// Create a slice of Container structs
	containers := []Container{}

	// Iterate through the containers and append them to the slice
	for _, containerInfo := range containersInfo {

		// Extract fields from the container info and map them to the Container struct
		containerID := containerInfo.ID[:12]
		taskOwner := containerInfo.Labels["TaskOwner"]
		targetName := containerInfo.Labels["Target"]
		vpnRegion := containerInfo.Labels["vpn.region"]
		vpnSOCKSPort := containerInfo.Labels["vpn.socks.port"]
		vpnHTTPPort := containerInfo.Labels["vpn.http.port"]
		url := fmt.Sprintf("/logs-viewer?container_id=%s", containerInfo.ID[:12])

		switch containerType {

		// If the container type is "scraper", only list scraper containers
		case "scraper":
			// logic for listing scraper containers
			if taskOwner != "" && taskOwner != "PROXY" {

				containers = append(containers, Container{
					ContainerID: &containerID,
					URL:         &url,
					TaskOwner:   &taskOwner,
					TargetName:  &targetName,
					VPNRegion:   &vpnRegion,
				})
			}

			// If the container type is "proxy", only list proxy containers
		case "proxy":
			if taskOwner == "PROXY" {
				containers = append(containers, Container{
					ContainerID:  &containerID,
					VPNRegion:    &vpnRegion,
					IPAddress:    &containerInfo.NetworkSettings.Networks["scraper_vpn"].IPAddress,
					VPNSOCKSPort: &vpnSOCKSPort,
					VPNHTTPPort:  &vpnHTTPPort,
				})

			}

		default:
			utils.ErrorHandler(fmt.Errorf("Invalid container type"))
		}
	}

	return containers
}

// ProxyContainer information
type ProxyContainer struct {
	ContainerID  string
	LockKey      string
	IPAddress    string
	VPNRegion    string
	VPNSOCKSPort string
	VPNHTTPPort  string
}

// AcquireProxyContainer acquires a lock on a proxy container and returns its ID
func AcquireProxyContainer() ProxyContainer {
	availableProxies := ListContainersByType("proxy")

	for _, proxy := range availableProxies {
		lockKey := "proxy-usage:" + *proxy.ContainerID
		lockSuccess := database.SetLock(lockKey)

		if lockSuccess {
			return ProxyContainer{
				ContainerID:  *proxy.ContainerID,
				LockKey:      lockKey,
				IPAddress:    *proxy.IPAddress,
				VPNRegion:    *proxy.VPNRegion,
				VPNSOCKSPort: *proxy.VPNSOCKSPort,
				VPNHTTPPort:  *proxy.VPNHTTPPort,
			}
		}
		// If the lock is not successful, try the next proxy container
	}

	// If no proxy container could be locked, return an empty string
	return ProxyContainer{}
}

// ReleaseProxyContainer releases the lock on a proxy container
func ReleaseProxyContainer(containerID string) {
	lockKey := "proxy-usage:" + containerID
	log.Printf("Releasing lock on proxy container %s", lockKey)
	database.ReleaseLock(lockKey)
}

// GetResultCSVSizeInContainer gets the size of the result csv file in the container
func getResultCSVSizeInContainer(containerID, filePathInContainer string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.ErrorHandler(err)
	defer cli.Close()

	// Log the file size in the container
	containerFileInfo, err := cli.ContainerStatPath(context.Background(), containerID, filePathInContainer)
	if err == nil {
		log.Printf("File size in container: %d bytes", containerFileInfo.Size)
	} else {
		log.Printf("Error getting file size in container: %v", err)
	}
}
