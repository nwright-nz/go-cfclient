package cfclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type DropletRequest struct {
	Data struct {
		GUID string `json:"guid,omitempty"`
	} `json:"data,omitempty"`
}

type AppResponse struct {
	Count     int           `json:"total_results"`
	Pages     int           `json:"total_pages"`
	NextUrl   string        `json:"next_url"`
	Resources []AppResource `json:"resources"`
}

type AppResource struct {
	Meta   Meta `json:"metadata"`
	Entity App  `json:"entity"`
}

type App struct {
	Guid                     string                 `json:"guid"`
	CreatedAt                string                 `json:"created_at"`
	UpdatedAt                string                 `json:"updated_at"`
	Name                     string                 `json:"name"`
	Memory                   int                    `json:"memory"`
	Instances                int                    `json:"instances"`
	DiskQuota                int                    `json:"disk_quota"`
	SpaceGuid                string                 `json:"space_guid"`
	StackGuid                string                 `json:"stack_guid"`
	State                    string                 `json:"state"`
	PackageState             string                 `json:"package_state"`
	Command                  string                 `json:"command"`
	Buildpack                string                 `json:"buildpack"`
	DetectedBuildpack        string                 `json:"detected_buildpack"`
	DetectedBuildpackGuid    string                 `json:"detected_buildpack_guid"`
	HealthCheckHttpEndpoint  string                 `json:"health_check_http_endpoint"`
	HealthCheckType          string                 `json:"health_check_type"`
	HealthCheckTimeout       int                    `json:"health_check_timeout"`
	Diego                    bool                   `json:"diego"`
	EnableSSH                bool                   `json:"enable_ssh"`
	DetectedStartCommand     string                 `json:"detected_start_command"`
	DockerImage              string                 `json:"docker_image"`
	DockerCredentials        map[string]interface{} `json:"docker_credentials_json"`
	Environment              map[string]interface{} `json:"environment_json"`
	StagingFailedReason      string                 `json:"staging_failed_reason"`
	StagingFailedDescription string                 `json:"staging_failed_description"`
	Ports                    []int                  `json:"ports"`
	SpaceURL                 string                 `json:"space_url"`
	SpaceData                SpaceResource          `json:"space"`
	PackageUpdatedAt         string                 `json:"package_updated_at"`
	c                        *Client
}

type V3DockerApp struct {
	Name                 string            `json:"name"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	Relationships        struct {
		Space struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"space"`
	} `json:"relationships"`
	Lifecycle struct {
		Type string `json:"type"`
		Data struct {
		} `json:"data"`
	} `json:"lifecycle"`
}

type V3DockerPackage struct {
	Type          string `json:"type"`
	Relationships struct {
		App struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"app"`
	} `json:"relationships"`
	Data struct {
		Image string `json:"image"`
	} `json:"data"`
}

type V3DockerPackageResponse struct {
	GUID      string `json:"guid"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type V3DockerBuildResponse struct {
	GUID      string `json:"guid"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Droplet   struct {
		GUID string `json:"guid"`
	} `json:"droplet"`
}

type V3DockerBuild struct {
	Package struct {
		GUID string `json:"guid"`
	} `json:"package"`
}

type V3DockerAppResponse struct {
	GUID      string `json:"guid"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Lifecycle struct {
		Type string `json:"type"`
		Data struct {
		}
	}
}

type AppInstance struct {
	State string    `json:"state"`
	Since sinceTime `json:"since"`
}

type AppStats struct {
	State string `json:"state"`
	Stats struct {
		Name      string   `json:"name"`
		Uris      []string `json:"uris"`
		Host      string   `json:"host"`
		Port      int      `json:"port"`
		Uptime    int      `json:"uptime"`
		MemQuota  int      `json:"mem_quota"`
		DiskQuota int      `json:"disk_quota"`
		FdsQuota  int      `json:"fds_quota"`
		Usage     struct {
			Time statTime `json:"time"`
			CPU  float64  `json:"cpu"`
			Mem  int      `json:"mem"`
			Disk int      `json:"disk"`
		} `json:"usage"`
	} `json:"stats"`
}

type AppSummary struct {
	Guid                     string                 `json:"guid"`
	Name                     string                 `json:"name"`
	ServiceCount             int                    `json:"service_count"`
	RunningInstances         int                    `json:"running_instances"`
	SpaceGuid                string                 `json:"space_guid"`
	StackGuid                string                 `json:"stack_guid"`
	Buildpack                string                 `json:"buildpack"`
	DetectedBuildpack        string                 `json:"detected_buildpack"`
	Environment              map[string]interface{} `json:"environment_json"`
	Memory                   int                    `json:"memory"`
	Instances                int                    `json:"instances"`
	DiskQuota                int                    `json:"disk_quota"`
	State                    string                 `json:"state"`
	Command                  string                 `json:"command"`
	PackageState             string                 `json:"package_state"`
	HealthCheckType          string                 `json:"health_check_type"`
	HealthCheckTimeout       int                    `json:"health_check_timeout"`
	StagingFailedReason      string                 `json:"staging_failed_reason"`
	StagingFailedDescription string                 `json:"staging_failed_description"`
	Diego                    bool                   `json:"diego"`
	DockerImage              string                 `json:"docker_image"`
	DetectedStartCommand     string                 `json:"detected_start_command"`
	EnableSSH                bool                   `json:"enable_ssh"`
	DockerCredentials        map[string]interface{} `json:"docker_credentials_json"`
}

type AppEnv struct {
	// These can have arbitrary JSON so need to map to interface{}
	Environment    map[string]interface{} `json:"environment_json"`
	StagingEnv     map[string]interface{} `json:"staging_env_json"`
	RunningEnv     map[string]interface{} `json:"running_env_json"`
	SystemEnv      map[string]interface{} `json:"system_env_json"`
	ApplicationEnv map[string]interface{} `json:"application_env_json"`
}

// Custom time types to handle non-RFC3339 formatting in API JSON

type sinceTime struct {
	time.Time
}

func (s *sinceTime) UnmarshalJSON(b []byte) (err error) {
	timeFlt, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	time := time.Unix(int64(timeFlt), 0)
	*s = sinceTime{time}
	return nil
}

func (s sinceTime) ToTime() time.Time {
	t, _ := time.Parse(time.UnixDate, s.Format(time.UnixDate))
	return t
}

type statTime struct {
	time.Time
}

func (s *statTime) UnmarshalJSON(b []byte) (err error) {
	timeString, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	possibleFormats := [...]string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05 -0700", "2006-01-02 15:04:05 MST"}

	var value time.Time
	for _, possibleFormat := range possibleFormats {
		if value, err = time.Parse(possibleFormat, timeString); err == nil {
			*s = statTime{value}
			return nil
		}
	}

	return fmt.Errorf("%s was not in any of the expected Date Formats %v", timeString, possibleFormats)
}

func (s statTime) ToTime() time.Time {
	t, _ := time.Parse(time.UnixDate, s.Format(time.UnixDate))
	return t
}

func (a *App) Space() (Space, error) {
	var spaceResource SpaceResource
	r := a.c.NewRequest("GET", a.SpaceURL)
	resp, err := a.c.DoRequest(r)
	if err != nil {
		return Space{}, errors.Wrap(err, "Error requesting space")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Space{}, errors.Wrap(err, "Error reading space response")
	}

	err = json.Unmarshal(resBody, &spaceResource)
	if err != nil {
		return Space{}, errors.Wrap(err, "Error unmarshalling body")
	}
	spaceResource.Entity.Guid = spaceResource.Meta.Guid
	spaceResource.Entity.c = a.c
	return spaceResource.Entity, nil
}

// ListAppsByQueryWithLimits queries totalPages app info. When totalPages is
// less and equal than 0, it queries all app info
// When there are no more than totalPages apps on server side, all apps info will be returned
func (c *Client) ListAppsByQueryWithLimits(query url.Values, totalPages int) ([]App, error) {
	return c.listApps("/v2/apps?"+query.Encode(), totalPages)
}

func (c *Client) ListAppsByQuery(query url.Values) ([]App, error) {
	return c.listApps("/v2/apps?"+query.Encode(), -1)
}

func (c *Client) ListApps() ([]App, error) {
	q := url.Values{}
	q.Set("inline-relations-depth", "2")
	return c.ListAppsByQuery(q)
}

func (c *Client) ListAppsByRoute(routeGuid string) ([]App, error) {
	return c.listApps(fmt.Sprintf("/v2/routes/%s/apps", routeGuid), -1)
}

func (c *Client) listApps(requestUrl string, totalPages int) ([]App, error) {
	pages := 0
	apps := []App{}
	for {
		var appResp AppResponse
		r := c.NewRequest("GET", requestUrl)
		resp, err := c.DoRequest(r)
		if err != nil {
			return nil, errors.Wrap(err, "Error requesting apps")
		}
		defer resp.Body.Close()
		resBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "Error reading app request")
		}

		err = json.Unmarshal(resBody, &appResp)
		if err != nil {
			return nil, errors.Wrap(err, "Error unmarshalling app")
		}

		for _, app := range appResp.Resources {
			app.Entity.Guid = app.Meta.Guid
			app.Entity.CreatedAt = app.Meta.CreatedAt
			app.Entity.UpdatedAt = app.Meta.UpdatedAt
			app.Entity.SpaceData.Entity.Guid = app.Entity.SpaceData.Meta.Guid
			app.Entity.SpaceData.Entity.OrgData.Entity.Guid = app.Entity.SpaceData.Entity.OrgData.Meta.Guid
			app.Entity.c = c
			apps = append(apps, app.Entity)
		}

		requestUrl = appResp.NextUrl
		if requestUrl == "" {
			break
		}

		pages += 1
		if totalPages > 0 && pages >= totalPages {
			break
		}
	}
	return apps, nil
}

func (c *Client) GetAppInstances(guid string) (map[string]AppInstance, error) {
	var appInstances map[string]AppInstance

	requestURL := fmt.Sprintf("/v2/apps/%s/instances", guid)
	r := c.NewRequest("GET", requestURL)
	resp, err := c.DoRequest(r)
	if err != nil {
		return nil, errors.Wrap(err, "Error requesting app instances")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading app instances")
	}
	err = json.Unmarshal(resBody, &appInstances)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshalling app instances")
	}
	return appInstances, nil
}

func (c *Client) GetAppEnv(guid string) (AppEnv, error) {
	var appEnv AppEnv

	requestURL := fmt.Sprintf("/v2/apps/%s/env", guid)
	r := c.NewRequest("GET", requestURL)
	resp, err := c.DoRequest(r)
	if err != nil {
		return appEnv, errors.Wrap(err, "Error requesting app env")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return appEnv, errors.Wrap(err, "Error reading app env")
	}
	err = json.Unmarshal(resBody, &appEnv)
	if err != nil {
		return appEnv, errors.Wrap(err, "Error unmarshalling app env")
	}
	return appEnv, nil
}

func (c *Client) GetAppRoutes(guid string) ([]Route, error) {
	return c.fetchRoutes(fmt.Sprintf("/v2/apps/%s/routes", guid))
}

func (c *Client) GetAppStats(guid string) (map[string]AppStats, error) {
	var appStats map[string]AppStats

	requestURL := fmt.Sprintf("/v2/apps/%s/stats", guid)
	r := c.NewRequest("GET", requestURL)
	resp, err := c.DoRequest(r)
	if err != nil {
		return nil, errors.Wrap(err, "Error requesting app stats")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading app stats")
	}
	err = json.Unmarshal(resBody, &appStats)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshalling app stats")
	}
	return appStats, nil
}

func (c *Client) KillAppInstance(guid string, index string) error {
	requestURL := fmt.Sprintf("/v2/apps/%s/instances/%s", guid, index)
	r := c.NewRequest("DELETE", requestURL)
	resp, err := c.DoRequest(r)
	if err != nil {
		return errors.Wrapf(err, "Error stopping app %s at index %s", guid, index)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return errors.Wrapf(err, "Error stopping app %s at index %s", guid, index)
	}
	return nil
}

func (c *Client) GetAppByGuid(guid string) (App, error) {
	var appResource AppResource
	r := c.NewRequest("GET", "/v2/apps/"+guid+"?inline-relations-depth=2")
	resp, err := c.DoRequest(r)
	if err != nil {
		return App{}, errors.Wrap(err, "Error requesting apps")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return App{}, errors.Wrap(err, "Error reading app response body")
	}

	err = json.Unmarshal(resBody, &appResource)
	if err != nil {
		return App{}, errors.Wrap(err, "Error unmarshalling app")
	}
	appResource.Entity.Guid = appResource.Meta.Guid
	appResource.Entity.SpaceData.Entity.Guid = appResource.Entity.SpaceData.Meta.Guid
	appResource.Entity.SpaceData.Entity.OrgData.Entity.Guid = appResource.Entity.SpaceData.Entity.OrgData.Meta.Guid
	appResource.Entity.c = c
	return appResource.Entity, nil
}

func (c *Client) AppByGuid(guid string) (App, error) {
	return c.GetAppByGuid(guid)
}

//CreateV3DockerBuild creates a build to stage the docker image. Needs to be associated
//with an existing package
func (c *Client) CreateV3DockerBuild(pkgGUID string) (bld V3DockerBuildResponse, err error) {
	v3Build := V3DockerBuild{}
	v3Build.Package.GUID = pkgGUID
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(v3Build)
	r := c.NewRequestWithBody("POST", "/v3/builds", b)
	resp, err := c.DoRequest(r)

	if err != nil {
		return bld, errors.Wrap(err, "Error requesting build")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bld, errors.Wrap(err, "Error reading build response")
	}

	err = json.Unmarshal(resBody, &bld)
	if err != nil {
		return bld, errors.Wrap(err, "Error unmarshalling build response")
	}
	return bld, nil

}

//CreateV3DockerPackage creates the required package for an application object
//It then needs a build associated.
func (c *Client) CreateV3DockerPackage(appGUID, image string) (pkg V3DockerPackageResponse, err error) {
	v3Package := V3DockerPackage{}
	v3Package.Data.Image = image
	v3Package.Relationships.App.Data.GUID = appGUID
	v3Package.Type = "docker"
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(v3Package)
	r := c.NewRequestWithBody("POST", "/v3/packages", b)
	resp, err := c.DoRequest(r)

	if err != nil {
		return pkg, errors.Wrap(err, "Error requesting package creation")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return pkg, errors.Wrap(err, "Error reading package response")
	}

	err = json.Unmarshal(resBody, &pkg)
	if err != nil {
		return pkg, errors.Wrap(err, "Error unmarshalling package response")
	}
	return pkg, nil
}

func (c *Client) AssignDropletToApp(appGUID, dropletGUID string) (V3DockerAppResponse, error) {
	app := V3DockerAppResponse{}
	droplet := DropletRequest{}
	droplet.Data.GUID = dropletGUID

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(droplet)
	r := c.NewRequestWithBody("PATCH", "/v3/apps/"+appGUID+"/relationships/current_droplet", b)
	resp, err := c.DoRequest(r)
	fmt.Println(r)

	if err != nil {
		fmt.Println(err)
		return V3DockerAppResponse{}, errors.Wrap(err, "Error requesting droplet")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return V3DockerAppResponse{}, errors.Wrap(err, "Error reading app response")
	}

	err = json.Unmarshal(resBody, &app)
	if err != nil {
		return V3DockerAppResponse{}, errors.Wrap(err, "Error unmarshalling app response")
	}
	return V3DockerAppResponse{}, nil

}

//GetV3DockerBuild checks to see if a given build has successfulyl staged, so it can
//get the droplet guid, and apply it to the application
func (c *Client) GetV3BuildInfo(bldGUID string) (bld V3DockerBuildResponse, err error) {
	r := c.NewRequest("GET", "/v3/builds/"+bldGUID)
	resp, err := c.DoRequest(r)
	if err != nil {
		return bld, errors.Wrap(err, "Error getting build information")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bld, errors.Wrap(err, "Error reading build response")
	}

	err = json.Unmarshal(resBody, &bld)
	if err != nil {
		return bld, errors.Wrap(err, "Error unmarshalling build body")
	}
	return bld, nil

}

//CreateV3DockerApp takes an appname, and a GUID for space. It will then
//create the app object and return the GUID.
func (c *Client) CreateV3DockerApp(appName, spaceGuid string) (app V3DockerAppResponse, err error) {

	v3AppSpec := V3DockerApp{}
	v3AppSpec.Name = appName
	v3AppSpec.Relationships.Space.Data.GUID = spaceGuid
	v3AppSpec.Lifecycle.Type = "docker"
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(v3AppSpec)
	r := c.NewRequestWithBody("POST", "/v3/apps", b)
	resp, err := c.DoRequest(r)
	if err != nil {
		return app, errors.Wrap(err, "Error requesting app env")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return app, errors.Wrap(err, "Error reading app env")
	}

	err = json.Unmarshal(resBody, &app)
	if err != nil {
		return app, errors.Wrap(err, "Error unmarshalling app env")
	}
	return app, nil

}

func (c *Client) StartApp(appGUID string) (app V3DockerAppResponse, err error) {
	cfUpdateRequest := c.NewRequest("POST", "/v3/apps/"+appGUID+"/actions/start")
	_, err = c.DoRequest(cfUpdateRequest)
	if err != nil {
		fmt.Println(err)
	}

	return app, err
}

//CreateV3DockerAppWithEnv takes an appname, a guid for the space and a map of strings for the environment vars.
// It will then create the app object
func (c *Client) CreateV3DockerAppWithEnv(appName, spaceGuid string, vars map[string]string) (app V3DockerAppResponse, err error) {

	v3AppSpec := V3DockerApp{}
	v3AppSpec.Name = appName
	v3AppSpec.Relationships.Space.Data.GUID = spaceGuid
	v3AppSpec.Lifecycle.Type = "docker"
	v3AppSpec.EnvironmentVariables = vars
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(v3AppSpec)
	r := c.NewRequestWithBody("POST", "/v3/apps", b)
	resp, err := c.DoRequest(r)
	if err != nil {
		return app, errors.Wrap(err, "Error requesting app env")
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return app, errors.Wrap(err, "Error reading app env")
	}

	err = json.Unmarshal(resBody, &app)
	if err != nil {
		return app, errors.Wrap(err, "Error unmarshalling app env")
	}
	return app, nil

}

//AppByName takes an appName, and GUIDs for a space and org, and performs
// the API lookup with those query parameters set to return you the desired
// App object.
func (c *Client) AppByName(appName, spaceGuid, orgGuid string) (app App, err error) {
	query := url.Values{}
	query.Add("q", fmt.Sprintf("organization_guid:%s", orgGuid))
	query.Add("q", fmt.Sprintf("space_guid:%s", spaceGuid))
	query.Add("q", fmt.Sprintf("name:%s", appName))
	apps, err := c.ListAppsByQuery(query)
	if err != nil {
		return
	}
	if len(apps) == 0 {
		err = fmt.Errorf("No app found with name: `%s` in space with GUID `%s` and org with GUID `%s`", appName, spaceGuid, orgGuid)
		return
	}
	app = apps[0]
	return
}
