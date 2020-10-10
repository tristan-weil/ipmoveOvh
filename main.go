package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ovh/go-ovh/ovh"
)

type (
	partialService struct {
		Description string `json:"description"`
		ProjectId   string `json:"project_id"`
	}

	partialIpFailOver struct {
		Id       string `json:"id"`
		Ip       string `json:"ip"`
		RoutedTo string `json:"routedTo"`
		Status   string `json:"status"`
	}

	partialInstance struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	projectIpFailoverAttachCreation struct {
		InstanceId string `json:"instanceId"`
	}
)

func main() {
	//
	// command line
	//
	logLevelPtr := flag.String("log-level", "info", `The log level [info|debug]`)

	myServicePtr := flag.String("project", "", `The name of your OVH project/service`)
	myIpPtr := flag.String("ip", "", `The FailOver IP to move`)
	myInstancePtr := flag.String("instance", "", `The name of the instance to attach the FailOver IP to`)
	myServiceIdPtr := flag.String("project-id", "", `The id of your OVH project/service.`)
	myIpIdPtr := flag.String("ip-id", "", `The id of the FailOver IP to move`)
	myInstanceIdPtr := flag.String("instance-id", "", `The id of the instance to attach the FailOver IP to`)
	flag.Parse()

	logLevel := *logLevelPtr
	myService := *myServicePtr
	myIp := *myIpPtr
	myInstance := *myInstancePtr
	myServiceId := *myServiceIdPtr
	myIpId := *myIpIdPtr
	myInstanceId := *myInstanceIdPtr

	log.SetFlags(0)

	if myService == "" && myServiceId == "" {
		log.Fatalf("FATAL: The parameter `service' or `servce-id' must be defined")
	}
	if myIp == "" && myIpId == "" {
		log.Fatalf("FATAL: The parameter `failover-ip' or `failover-ip-id' must be defined")
	}
	if myInstance == "" && myInstanceId == "" {
		log.Fatalf("FATAL: The parameter `instance' or `instance-id' must be defined")
	}

	//
	// login
	//
	if logLevel == "debug" {
		log.Printf("DEBUG: Creating the connection to OVH API")
	}
	client, err := ovh.NewDefaultClient()
	if err != nil {
		log.Fatalf("FATAL: Unable to create a connnection to OVH API: %s", err)
	}

	//
	// get service
	//
	var service *partialService

	if myServiceId != "" {
		// check the service id exists
		if logLevel == "debug" {
			log.Printf("DEBUG: Getting the project/service")
			log.Printf("DEBUG: GET /cloud/project/%s", myServiceId)
		}

		err = client.Get(fmt.Sprintf("/cloud/project/%s", myServiceId), &service)
		if err != nil {
			log.Fatalf("FATAL: Unable to the project/service Id '%s': `%s'", myServiceId, err)
		}

	} else {
		// get list of services
		if logLevel == "debug" {
			log.Printf("DEBUG: Getting the list of projects/services")
			log.Printf("DEBUG: GET /cloud/project")
		}

		var servicesIds []string

		err = client.Get("/cloud/project", &servicesIds)
		if err != nil {
			log.Fatalf("FATAL: Unable to get the list of services: `%s'", err)
		}

		// find our service
		if logLevel == "debug" {
			log.Printf("DEBUG: Finding the project/service '%s' in the list of projects/services", myService)
		}

		for _, serviceId := range servicesIds {
			if logLevel == "debug" {
				log.Printf("DEBUG: GET /cloud/project/%s", serviceId)
			}

			var curService partialService
			err = client.Get(fmt.Sprintf("/cloud/project/%s", serviceId), &curService)
			if err != nil {
				log.Fatalf("FATAL: Unable to get information about the project/service '%s': `%s'", serviceId, err)
			}

			if curService.Description == myService {
				if logLevel == "debug" {
					log.Printf("DEBUG: Found!")
				}

				service = &curService
				break
			}
		}

		if service == nil {
			log.Fatalf("FATAL: Unable to find the service '%s' in the list of projects/services", myService)
		}
	}

	//
	// get the failover-ip
	//
	var ip *partialIpFailOver

	if myIpId != "" {
		// check the failover-ip id exists
		if logLevel == "debug" {
			log.Printf("DEBUG: Getting the failover ip")
			log.Printf("DEBUG: GET /cloud/project/%s/ip/failover/%s", service.ProjectId, myIpId)
		}

		err = client.Get(fmt.Sprintf("/cloud/project/%s/ip/failover/%s", service.ProjectId, myIpId), &ip)
		if err != nil {
			log.Fatalf("FATAL: Unable to find the FailOver IP id '%s' for the project/service '%s': `%s'", myIpId, service.Description, err)
		}
	} else {
		// get list of failover-ips
		if logLevel == "debug" {
			log.Printf("DEBUG: Getting the list of FailOver IPs")
			log.Printf("DEBUG: GET /cloud/project/%s/ip/failover", service.ProjectId)
		}

		var ips []partialIpFailOver
		err = client.Get(fmt.Sprintf("/cloud/project/%s/ip/failover", service.ProjectId), &ips)
		if err != nil {
			log.Fatalf("FATAL: Unable to get the list of FailOver IPs for the project/service '%s': `%s'", service.Description, err)
		}

		// find our ip
		if logLevel == "debug" {
			log.Printf("DEBUG: Finding the FailOver IP '%s' in the list of FailOver IPs", myIp)
		}

		for _, curIp := range ips {
			if curIp.Ip == myIp {
				if logLevel == "debug" {
					log.Printf("DEBUG: Found!")
				}

				ip = &curIp
				break
			}
		}

		if ip == nil {
			log.Fatalf("FATAL: Unable to find the FailOver IP '%s' in the list of FailOver IPs for the project/service '%s'", myIp, service.Description)
		}
	}

	//
	// get the instance
	//
	var instance *partialInstance

	if myInstanceId != "" {
		// check the instance id exists
		if logLevel == "debug" {
			log.Printf("DEBUG: Getting the instance")
			log.Printf("DEBUG: GET /cloud/project/%s/instance/%s", service.ProjectId, myInstanceId)
		}

		err = client.Get(fmt.Sprintf("/cloud/project/%s/instance/%s", service.ProjectId, myInstanceId), &instance)
		if err != nil {
			log.Fatalf("FATAL: Unable to find the instance id '%s' for the project/service '%s': `%s'", myInstanceId, service.Description, err)
		}
	} else {
		// get list of instances
		if logLevel == "debug" {
			log.Printf("DEBUG: Getting the list of instances")
			log.Printf("DEBUG: GET /cloud/project/%s/instance", service.ProjectId)
		}

		var instances []partialInstance
		err = client.Get(fmt.Sprintf("/cloud/project/%s/instance", service.ProjectId), &instances)
		if err != nil {
			log.Fatalf("FATAL: Unable to get the list of instances for project/service '%s': `%s'", service.Description, err)
		}

		// find our instance
		if logLevel == "debug" {
			log.Printf("DEBUG: Finding the instance '%s' in the list of instances", myInstance)
		}

		for _, curInstance := range instances {
			if curInstance.Name == myInstance {
				if logLevel == "debug" {
					log.Printf("DEBUG: Found!")
				}

				instance = &curInstance
				break
			}
		}

		if instance == nil {
			log.Fatalf("FATAL: Unable to find the instance '%s' in the list of instances for the project/service '%s'", myInstance, service.Description)
		}
	}

	//
	// check not already attached
	//
	if logLevel == "debug" {
		log.Printf("DEBUG: Getting the current instance the FailOver IP '%s' is attached to", ip.Ip)
		log.Printf("DEBUG: GET /cloud/project/%s/instance/%s", service.ProjectId, ip.RoutedTo)
	}

	var prevInstance *partialInstance

	err = client.Get(fmt.Sprintf("/cloud/project/%s/instance/%s", service.ProjectId, ip.RoutedTo), &prevInstance)
	if err != nil {
		log.Fatalf("FATAL: Unable to get the current instance the FailOver IP is attached to for project/service '%s': `%s'", service.Description, err)
	}

	if logLevel == "debug" {
		log.Printf("DEBUG: FailOver IP is currently attached to: %s (%s)", prevInstance.Name, prevInstance.Id)
		log.Printf("DEBUG: FailOver IP will be attached to: %s (%s)", instance.Name, instance.Id)
	}

	if ip.RoutedTo == instance.Id {
		log.Printf("The FailOver IP '%s' is already routed to '%s'", ip.Ip, instance.Name)
		os.Exit(0)
	}

	//
	// attach
	//
	if logLevel == "debug" {
		log.Printf("DEBUG: Attaching the FailOver IP '%s' to the instance", ip.Ip)
		log.Printf("DEBUG: POST /cloud/project/%s/ip/failover/%s/attach", service.ProjectId, ip.Id)
	}

	reqBody := projectIpFailoverAttachCreation{
		InstanceId: instance.Id,
	}

	// request
	err = client.Post(fmt.Sprintf("/cloud/project/%s/ip/failover/%s/attach", service.ProjectId, ip.Id), reqBody, ip)
	if err != nil {
		log.Fatalf("FATAL: Unable to attach the FailOver IP '%s' to the instance '%s' for project/server '%s': `%s'", ip.Ip, instance.Name, service.Description, err)
	}

	// check status
	for {
		if logLevel == "debug" {
			log.Printf("DEBUG: Status of the FailOver IP (waiting for 'ok'): %s", ip.Status)
		}

		if ip.Status != "ok" {
			if logLevel == "debug" {
				log.Printf("DEBUG: POST /cloud/project/%s/ip/failover/%s/attach", service.ProjectId, ip.Id)
			}

			err = client.Get(fmt.Sprintf("/cloud/project/%s/ip/failover/%s", service.ProjectId, ip.Id), ip)
			if err != nil {
				log.Fatalf("FATAL: Unable to attach the FailOver IP '%s' to the instance '%s' for project/server '%s': `%s'", ip.Ip, instance.Name, service.Description, err)
			}

			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	log.Printf("The FailOver IP '%s' is now attached to the instance '%s' (previous instance was '%s')", ip.Ip, instance.Name, prevInstance.Name)
	os.Exit(0)
}
