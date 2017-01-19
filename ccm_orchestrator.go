package main

import (
	"encoding/json"
	"flag"
	"github.com/misale/circonus-ccm/ccm"
	"github.com/misale/circonus-api-go/circonusapi"
	"log"
	"os"
)

var ccm_path = flag.String("repo", os.Getenv("CCM_REPO_PATH"), "Path to ccm repository directory")

func main() {
	// parse input flags
	flag.Parse()

	// set CCM_REPO_PATH
	os.Setenv("CCM_REPO_PATH", *ccm_path)
	tenant_info := ccm.TenantInfo()
	os.Setenv("CIRONUS_API_TOKEN", tenant_info.API_TOKEN)
	os.Setenv("CIRCONUS_APP_NAME", tenant_info.APP_NAME)
	os.Setenv("CIRCONUS_API_URL", tenant_info.API_URL)

	log.Println("[ ccm-orchestrator ] environmental variables set")
	log.Println("APP_NAME : ", tenant_info.APP_NAME)
	log.Println("API_TOKEN : ", tenant_info.API_TOKEN)
	log.Println("API_URL : ", tenant_info.API_URL)
	//////////////////////////////////////////////////////////////////////////////////////////////////////////

	/*	Main Logic
		1.Get Config file List
		 For each Config file: get template_file, get host list
		  - For each host Generate new check_bundle blob - GO Routin

		2.Circonus API search for the check_bundle using : target,type,display_name - Go Routin
		 If check_bundle doesnt exist create
	*/

	//	Configuration Files in path
	conf_file_list, err := ccm.ConfigFileList()
	ccm.ERRHandler(err, "ConfigFileList : ")
	log.Println("-> parsing ccm_configs\n", conf_file_list)
	var check_pipeline []<-chan ccm.CheckInput
	//
	for _, conf_file := range conf_file_list {
		log.Println("mapping configuration for: ", conf_file)
		ccm_conf := ccm.CCMRead(conf_file)
		template_file := os.Getenv("CCM_REPO_PATH") + "/ccm_templates/" + ccm_conf.TemplateFile
		configuration_file := os.Getenv("CCM_REPO_PATH") + "/ccm_configs/" + conf_file
		host_group := ccm_conf.HostGroup
		log.Println("host_group : ", host_group)
		hosts, err := ccm.GroupHostList(host_group)
		ccm.ERRHandler(err, host_group+" GroupHostList : ")
		log.Println("member hosts : ", hosts)
		for _, host := range hosts {
			log.Println("loading template: \t", template_file, " for host:\t", host)
			check_pipe := ccm.ZipRoutin(template_file, configuration_file, host)
			check_pipeline = append(check_pipeline, check_pipe)
		}

	}

	for _, check_pipe := range check_pipeline {
		//work with check_pipe
		//test if check_bundle exists
		check_input := <-check_pipe
		check_filter := circonusapi.CheckBundleFilter{
			Target:      check_input.Target,
			Type:        check_input.Type,
			DisplayName: check_input.DisplayName}
		cns_result, err := circonusapi.GetCns(check_filter, "check_bundle")
		//fmt.Println("cns_result ** : ", string(cns_result))
		ccm.ERRHandler(err, "GetCns")
		var check_bundle_result []circonusapi.CheckBundle
		ccm.ERRHandler(json.Unmarshal(cns_result, &check_bundle_result), "check bundle result unmarshal")
		//fmt.Println("check_bundle_result : ", check_bundle_result)

		if len(check_bundle_result) > 0 {
			log.Println("existing check detected for : ", check_input.Target, "\ttype:\t", check_input.Type, "\tdisplay_name:\t", check_input.DisplayName)
		} else {
			log.Println("creating new check_bundle for : ", check_input.Target, "\ttype:\t", check_input.Type, "\tdisplay_name:\t", check_input.DisplayName)
			//create check_bundle if it doesnt exist
			new_check_data, err := json.Marshal(check_input)
			ccm.ERRHandler(err, "check_input marshal ")
			new_check_bundle, err := circonusapi.CreateCns(new_check_data, "check_bundle")
			ccm.ERRHandler(err, "create check bundle ")
			log.Println("[new check_bundle]\n", string(new_check_bundle))
		}

	}
}
