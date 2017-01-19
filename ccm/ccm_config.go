package ccm

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

func ERRHandler(err error, message string) {
	if err != nil {
		log.Println("eRROR @ : ", message, " :", err)
	}
}

func TenantInfo() circonusTenant {
	ccm_path := os.Getenv("CCM_REPO_PATH")
	tenant_file := ccm_path + "/tenant.yml"
	file_data, err := ioutil.ReadFile(tenant_file)
	ERRHandler(err, "read tenant.yml")
	var tenantData circonusTenant
	ERRHandler(yaml.Unmarshal(file_data, &tenantData), "unmarshal tenant.yml")
	return tenantData
}

// HostGroups functions returns a list of of HostGroup structs reading all YAML files under ccm_hosts directory
func HostGroups() ([]HostGroup, error) {
	ccm_path := os.Getenv("CCM_REPO_PATH")
	host_file_path := ccm_path + "/ccm_hosts/"
	file_data, err := ioutil.ReadDir(host_file_path)
	ERRHandler(err, "read ccm_hosts dir")
	if err != nil {
		return nil, err
	}
	var HostsData []HostGroup
	for _, file_info := range file_data {
		host_file := host_file_path + file_info.Name()
		file_data, err := ioutil.ReadFile(host_file)
		ERRHandler(err, "read host_file in ccm_hosts "+host_file)
		if err != nil {
			return nil, err
		}
		var host_group HostGroup
		ERRHandler(yaml.Unmarshal(file_data, &host_group), "unmarshal host_file "+host_file)
		HostsData = append(HostsData, host_group)

	}

	return HostsData, nil
}

// GroupHostList function takes a group name string and returns a list of hosts (string) that belong to the host group
func GroupHostList(group string) ([]string, error) {
	//default result not found
	result := []string{}
	//get all host groups
	host_groups, err := HostGroups()
	if err != nil {
		return nil, err
	}

	for _, host_group := range host_groups {
		if host_group.GroupName == group {
			result = host_group.Members
		}
	}
	return result, nil
}

//ConfigFilesList function returns a list of config files (files under <repo_name>/ccm_config)
func ConfigFileList() ([]string, error) {
	ccm_path := os.Getenv("CCM_REPO_PATH")
	config_file_path := ccm_path + "/ccm_configs/"
	//fmt.Println(config_file_path)
	file_data, err := ioutil.ReadDir(config_file_path)
	ERRHandler(err, "read ccm_configs dir")
	if err != nil {
		return nil, err
	}
	var ConfigFiles []string
	for _, file_info := range file_data {
		config_file := file_info.Name()
		ConfigFiles = append(ConfigFiles, config_file)
	}

	return ConfigFiles, nil
}

//CCMRead function reads a ccm config file and returns CCMConf struct
func CCMRead(conf_file string) CCMConf {
	ccm_path := os.Getenv("CCM_REPO_PATH")
	config_file_path := ccm_path + "/ccm_configs/" + conf_file
	file_data, err := ioutil.ReadFile(config_file_path)
	ERRHandler(err, "read ccm_config file ")
	if err != nil {
		return CCMConf{}
	}
	var ccm_configuration CCMConf
	ERRHandler(json.Unmarshal(file_data, &ccm_configuration), "CcmConf unmarshal : ")
	return ccm_configuration
}

//Zipper function takes all variables in the template file have their values set using the config_file
//values and returns the []byte representation of a check bundle

func Zipper(template_file string, config_file string, host string) CheckInput {

	template_data, err := ioutil.ReadFile(template_file)
	ERRHandler(err, "template_file_read ")

	config_data, err := ioutil.ReadFile(config_file)
	ERRHandler(err, "config_file_read ")

	var check_config CcmTemplate
	ERRHandler(json.Unmarshal(config_data, &check_config), "check_config unmarshal ")

	//update target field
	check_config.Target = host
	template_parser := template.Must(template.New("check_config").Parse(string(template_data)))

	zipped_data := new(bytes.Buffer)
	ERRHandler(template_parser.Execute(zipped_data, check_config), "template execute : ")

	//fmt.Println("\n",zipped_data.String(),"\n")

	var parsed_check CheckInput
	ERRHandler(json.Unmarshal(zipped_data.Bytes(), &parsed_check), "parsed_check template unmarshal : ")

	return parsed_check

}

//ZipRoutin functions launches goroutins that map config to template per host and returns a receive only channel
func ZipRoutin(template_file string, configuration_file string, host string) <-chan CheckInput {
	check_holder := make(chan CheckInput)
	go func() {
		check_input := Zipper(template_file, configuration_file, host)
		check_input.Target = host
		log.Println("inspecting check configuration\n", check_input)
		check_holder <- check_input
		close(check_holder)
	}()
	return check_holder
}
