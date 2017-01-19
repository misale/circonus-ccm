# circonus check configuration management 

Creating new checks for each individual target/host and maintaining the check configuration becomes a very tedious task as Circonus 
footprint grows. The manual configurations/setups will not scale and are prone to human error. Circonus has provided API interfaces 
for all of the configuration tasks associated with setting up new checks. There are 2 broad classes of check_bundles in Circonus. 
PUSH (metrics data sent by an active agent on a host/application source) and PULL (metrics data is queried by Circonus broker, passive 
data source).  

#
**1.** In the case of NAD agent host check and most of the **PUSH** checks, configuration management can be handled by including the NAD agent installation, 
configuration (Circonus One Step Installation) into existing infrastructure automation tools.

**Example:** Puppet module for NAD agent (COSI) installation
 
 ```puppet
class circonus_agent {

  file {['/usr/local',
  '/usr/local/ccm',
  '/usr/local/ccm/bin',]:

    ensure => directory,
    mode   => 0755,
    owner  => root,
    group  => root,
  }

  file {'/usr/local/ccm/bin/install_nad_agent.sh' :
    ensure  => present,
    owner   => root,
    group   => root,
    mode    => 755,
    source  => 'puppet:///modules/circonus_agent/install_nad_agent.sh',
    require => File['/usr/local/ccm/bin'],
    notify  => Exec['exec_nad_installer'],
  }

  exec {'exec_nad_installer':
    command => '/usr/local/ccm/bin/install_nad_agent.sh',
    logoutput => true,
    #onlyif => '',
  }

}
```
[COSI](https://github.com/circonus-labs/circonus-one-step-install/wiki) bash script (install_nad_agent.sh)

```
#!/bin/bash
# Installs NAD agent using Circonus COSI - https://github.com/circonus-labs/circonus-one-step-install/wiki
#

KEY="fdffdfdffd-d241-c13c-97e3-234567def456a"
APP="circonus:osi:0018fd34728dffff7890dff87326067a35a1be7b2e3545bdaf3458372cbff759"

# Check if installed
if [[ -f /opt/circonus/cosi/bin/cosi ]] ; then
	output=`/opt/circonus/cosi/bin/cosi check list --long --verify`
	if [[ "$output" =~ "OK" ]]; then
		echo "NAD agent is installed and check is OK. Skipping nad install .. "
		echo $output
		exit 0
	fi

fi
# Run cosi
/usr/bin/curl  -sSL https://onestep.circonus.com/install | bash  -s -- --key $KEY --app $APP
```
#

**2.** **PULL** check [types](https://login.circonus.com/user/docs/Data/CheckTypes) are setup on 
Circonus portal and all of their configuration data lives in Circonus. **circonus-ccm** project
is a configuration management tool for such check_bundles, whose configuration is setup and lives only in Circonus. 
circonus-ccm utilizes a directory as configuration repository for check configurations in combination with
 [Circonus API](https://login.circonus.com/resources/api) for automation of the configuration management.
 
 
**CCM Configuration Repository**

circonus-ccm repository is laid out in directory tree as shown in the example below. The root directory ("ccm_test" in the example)
is the name of the particular repository so it can have any name, but the sub-directories under it should always be as shown below. 
(Refer to the description for each sub-dir below). ccm repository can be defined and managed by any infrastructure configuration 
management tool (puppet, ANSIBLE). A puppet manifest example is included below, and it can be setup on a puppetized utility host to 
centrally manage check_bundles. 

Example

```
ccm_test/
├── ccm_configs
│   └── check_x1.json
├── ccm_hosts
│   └── hosts.yml
├── ccm_templates
│   └── check_x.json
└── tenant.yml
```
 - **ccm_test** : Config repo name. Repo should have sub-directories "ccm_configs", "ccm_hosts" and "ccm_templates". And tenant.yml file. 

 - **ccm_templates** : Contains list of JSON check template files that have "golang text/template" formatting for variable replacement. Strings with **"{{.Param}}"** 
format are template variables to be replaced by values from configuration files. Refer to the configuration-to-template key map table below for list of template variables.

Example:
```go
{
  "brokers": ["/broker/{{.Broker}}"],
  "config": {
    "header_host":"{{.Target}}",
    "http_version":"{{.ConfigHTTPVersion}}",
    "method":"{{.ConfigMethod}}",
    "payload":"{{.ConfigPayload}}",
    "port":"{{.ConfigPort}}",
    "read_limit":"{{.ConfigReadLimit}}",
    "url":"http://{{.Target}}/jobs/update_filter.json",
    "query":"{{.ConfigQuery}}"
  },
  "display_name": "{{.DisplayName}} {{.Target}}",
  "notes": "{{.Notes}}",
  "period": {{.Period}},
  "tags": ["tag:tag1","tag:tag2"],
  "target": "{{.Target}}",
  "timeout": {{.Timeout}},
  "type": "{{.Type}}",
  "metrics": [
      {
        "status": "active",
        "name": "latency",
        "type": "numeric",
        "units": null,
        "tags" : ["metric_tag:metric_tag1","metric_tag:metric_tag2"]
      },
      {
        "status": "active",
        "name": "thruput",
        "type": "numeric",
        "units": null,
        "tags": ["metric_tag:metric_tagx","metric_tag:metric_tagy"]
      }
]
}
```

 - **ccm_configs** : Contains list of JSON configuration files. The value of "host_group" field corresponds to the group_name key in ccm_hosts files.
The value of "template_file" corresponds to the template file under ccm_templates directory where the values in the configuration file will be mapped.

Example :
```json
{
  "host_group" : "web_app",
  "template_file" : "check_x.json",
  "broker" : "9898",
  "display_name" : "app_stats",
  "notes" : "foo",
  "period": 120,
  "target" : "",
  "timeout" : 60,
  "type" : "json",
  "config_header_host":"",
  "config_http_version":"1.1",
  "config_method":"GET",
  "config_payload":"",
  "config_port":"80",
  "config_read_limit":"0",
  "config_url":"/stats",
  "config_query":""
}
```
**Configuration-to-Template Key Map**

The template keys list below is limited to making ccm functional for check_bundles of type JSON pull, HTTPTrap 
and CAQL for this version of ccm. (needs to be expanded to cover all check types).

| Template Key         | Configuration Key | Circonus Check Bundle field | Note |
|:---------------------|:------------------|:----------------------------|:-----|
|{{.HostGroup}}        |host_group         | -                           | group_name value in ccm_hosts   |
|{{.TemplateFile}}     |template_file      | -                           | template file name under ccm_templates    |
|{{.Broker}}           |broker             | A broker cid number          | cid number of broker (only the number part without the "/broker/" string) |
|{{.DisplayName}}      |display_name       | display_name                 | The name of the check that will be displayed in the web interface.|
|{{.Notes}}            |notes              | notes                        | Notes about this bundle |
|{{.Period}}           |period             | period                       | The period between each time the check is made. |
|{{.Target}}           |target             | target                            | What the check you're installing will check. |
|{{.Timeout}}          |timeout            | timeout                           | The length of time in seconds before the check will timeout if no response is returned to the broker.|
|{{.Type}}             |type               | type                              | The type of check that this is. This has no default value and must be supplied for each check. |
|{{.ConfigHeaderHost}} |config_header_host | header_host (check_bundle config) | check_bundle "config" section header_host field |
|{{.ConfigHTTPVersion}}|config_http_version| http_version (check_bundle config)| check_bundle "config" section http_version field |
|{{.ConfigMethod}}     |config_method      | method (check_bundle config)      | check_bundle "config" section method field |
|{{.ConfigPayload}}    |config_payload     | payload (check_bundle config)     | check_bundle "config" section payload field |
|{{.ConfigPort}}       |config_port        | port (check_bundle config)        | check_bundle "config" section port field |
|{{.ConfigReadLimit}}  |config_read_limit  | read_limit (check_bundle config)  | check_bundle "config" section read_limit field | 
|{{.ConfigURL}}        |config_url         | url (check_bundle config)         | check_bundle "config" section url field |
|{{.ConfigQuery}}      |config_query       | query (check_bundle config)       | check_bundle "config" section query field|

**ccm_hosts** : Contains yml files formatted as a "group_name" field followed by 
"members" field that has list of hosts belonging to the group. In the example below "web_app" is a the group_name.
 ccm_orchestrator uses the value of group_name in config files under ccm_configs directory to match the corresponding host group under ccm_hosts directory.
 
Example :

```yaml
group_name: web_app
members:
 - web-lab-alfa.xyz.net
 - web-lab-beta.xyz.net
 - web-lab-gama.xyz.net

```

**tenant.yml** : Circonus account information file. This file contains the Circonus API endpoint and authentication token for 
verifying and creating check bundles.

Example :
```yaml
# tenant : circonus account config file
circonus_api_token : 4e79ff58-6b94-603f-f64f-9b2761a7ffff
circonus_app_name : tenantx
circonus_api_url : https://api.circonus.com/v2/
```

**circonus-ccm puppet**

Example : ccm-test puppet manifest

```puppet
class circonus_ccm {

  File {
    ensure  => present,
    owner   => root,
    group   => root,
    mode    => 755,
  }
  # ccm directory tree
  file {['/test/ccm_test',
       '/test/ccm_test/ccm_configs',
       '/test/ccm_test/ccm_templates',
       '/test/ccm_test/ccm_hosts',]:

    ensure => directory,
 
  }
  # circonus tenant info
  file {'/test/ccm_test/tenant.yml' :  
        source  => 'puppet:///modules/circonus_ccm/tenant.yml',
        require => File['/test/ccm_test'],
        notify  => Exec['ccm_orchestrator'],
      }
  # hosts file
  file {'/test/ccm_test/ccm_hosts/elemental_hosts.yml' :
      source  => 'puppet:///modules/circonus_ccm/elemental_hosts.yml',
      require => File['/test/ccm_test/ccm_hosts'],
      notify  => Exec['ccm_orchestrator'],
    }
  # config file
  file {'/test/ccm_test/ccm_configs/check_x1.json' :
    source  => 'puppet:///modules/circonus_ccm/check_x1.json',
    require => File['/test/ccm_test/ccm_configs'],
    notify  => Exec['ccm_orchestrator'],
  }
  # circonus template file
  file {'/test/ccm_test/ccm_templates/check_x.json' :
      source  => 'puppet:///modules/circonus_ccm/check_x.json',
      require => File['/test/ccm_test/ccm_configs'],
      notify  => Exec['ccm_orchestrator'],
    }
    
  exec {'ccm_orchestrator':
    command => '/test/ccm_test/ccm_orchestrator -repo /test/ccm_test/',
    logoutput => true,
    #onlyif => '',
  }

}

```

Example : 

# 

### CCM flow

ccm_orchestrator is the CCM flow control script (needs to be run periodically or on-demand), it takes a path to CCM configuration directory as an argument. 
ccm_orchestrator first checks if a check_bundle exists and creates one if it doesn't. ccm_orchestrator is added as exec resource in the puppet manifest example above. 
ccm_orchestrator doesn't make any change/update to existing check_bundles (current version). The ccm_orchestrator runs a check_bundle search using "display_name", "target" and "type" 
fields of a check_bundle. Therefore the combination of "display_name", "target" and "type" fields is assumed to make a unique check_bundle for using circonus-ccm.

```
$ ./ccm_orchestrator -help
Usage of ./ccm_orchestrator:
  -repo string
        Path to ccm repository directory
$ 
```

Example :

```
$ ccm_orchestrator -repo ccm_test/
2017/01/08 23:14:20 [ ccm-orchestrator ] environmental variables set
2017/01/08 23:14:20 APP_NAME :  tenantx
2017/01/08 23:14:20 API_TOKEN :  b5fffff7-3b2e-ffff-b707-a47fffff4f4
2017/01/08 23:14:20 API_URL :  https://api.circonus.com/v2/
2017/01/08 23:14:20 -> parsing ccm_configs
 [check_x1.json]
2017/01/08 23:14:20 mapping configuration for:  check_x1.json
2017/01/08 23:14:20 host_group :  web_app
2017/01/08 23:14:20 member hosts :  [web-lab-alfa.xyz.net web-lab-beta.xyz.net web-lab-gama.xyz.net]
2017/01/08 23:14:20 loading template:    ccm_test//ccm_templates/check_x.json  for host:         web-lab-alfa.xyz.net
2017/01/08 23:14:20 loading template:    ccm_test//ccm_templates/check_x.json  for host:         web-lab-beta.xyz.net
2017/01/08 23:14:20 loading template:    ccm_test//ccm_templates/check_x.json  for host:         web-lab-gama.xyz.net
2017/01/08 23:14:20 inspecting check configuration
 {[/broker/1289] {web-lab-alfa.xyz.net 1.1 GET  80 0 http://web-lab-alfa.xyz.net/jobs/update_filter.json } elemental_stats web-lab-alfa.xyz.net notes 120 [tag:tag1 tag:tag2] web-lab-alfa.xyz.net 60 json [{active active numeric <nil> [metric_tag:metric_tag1 metric_tag:metric_tag2]} {active archived numeric <nil> [metric_tag:metric_tagx metric_tag:metric_tagy]}]}
2017/01/08 23:14:20 inspecting check configuration
 {[/broker/1289] {web-lab-gama.xyz.net 1.1 GET  80 0 http://web-lab-gama.xyz.net/jobs/update_filter.json } elemental_stats web-lab-gama.xyz.net notes 120 [tag:tag1 tag:tag2] web-lab-gama.xyz.net 60 json [{active active numeric <nil> [metric_tag:metric_tag1 metric_tag:metric_tag2]} {active archived numeric <nil> [metric_tag:metric_tagx metric_tag:metric_tagy]}]}
2017/01/08 23:14:20 inspecting check configuration
 {[/broker/1289] {web-lab-beta.xyz.net 1.1 GET  80 0 http://web-lab-beta.xyz.net/jobs/update_filter.json } elemental_stats web-lab-beta.xyz.net notes 120 [tag:tag1 tag:tag2] web-lab-beta.xyz.net 60 json [{active active numeric <nil> [metric_tag:metric_tag1 metric_tag:metric_tag2]} {active archived numeric <nil> [metric_tag:metric_tagx metric_tag:metric_tagy]}]}
2017/01/08 23:14:20 existing check detected for :  web-lab-alfa.xyz.net        type:    json   display_name:    elemental_stats web-lab-alfa.xyz.net
2017/01/08 23:14:20 existing check detected for :  web-lab-beta.xyz.net        type:    json   display_name:    elemental_stats web-lab-beta.xyz.net
2017/01/08 23:14:21 existing check detected for :  web-lab-gama.xyz.net        type:    json   display_name:    elemental_stats web-lab-gama.xyz.net
$ 
```
#
### Future work

- update/delete check_bundles based on changes in ccm config repo
- detect metrics changes within check_bundle and update check_bundles
- use uniform formatting for all ccm repo files - use YAML

