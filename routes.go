package cfclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/url"

	"github.com/pkg/errors"
)

type RoutesResponse struct {
	Count     int              `json:"total_results"`
	Pages     int              `json:"total_pages"`
	NextUrl   string           `json:"next_url"`
	Resources []RoutesResource `json:"resources"`
}

type RoutesResource struct {
	Meta struct {
		GUID string `json:"guid"`
	} `json:"metadata"`
	Entity Route `json:"entity"`
}

type RouteRequest struct {
	DomainGuid string `json:"domain_guid"`
	SpaceGuid  string `json:"space_guid"`
	Host       string `json:"host,omitempty"`
}

type RouteMap struct {
	AppGUID   string `json:"app_guid"`
	RouteGUID string `json:"route_guid"`
}

type MappedRoute struct {
	Metadata struct {
		GUID      string `json:"guid"`
		URL       string `json:"url"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	} `json:"metadata"`
	Entity struct {
		Port int `json:"app_port"`
	}
}
type Route struct {
	Guid                string `json:"guid"`
	Host                string `json:"host"`
	Path                string `json:"path"`
	DomainGuid          string `json:"domain_guid"`
	SpaceGuid           string `json:"space_guid"`
	ServiceInstanceGuid string `json:"service_instance_guid"`
	Port                int    `json:"port"`
	c                   *Client
}

func (c *Client) CreateTcpRoute(routeRequest RouteRequest) (Route, error) {
	routesResource, err := c.createRoute("/v2/routes?generate_port=true", routeRequest)
	if nil != err {
		return Route{}, err
	}
	return routesResource.Entity, nil
}

func (c *Client) CreateHttpRoute(routeRequest RouteRequest) (RoutesResource, error) {
	routesResource, err := c.createRoute("/v2/routes", routeRequest)
	if nil != err {
		return RoutesResource{}, err
	}
	return routesResource, nil
}

func (c *Client) MapRoute(routeMap RouteMap) (mr MappedRoute, err error) {
	buf := bytes.NewBuffer(nil)
	jsonErr := json.NewEncoder(buf).Encode(routeMap)
	requestURL := "/v2/route_mappings"
	if jsonErr != nil {
		return mr, errors.Wrap(err, "Error mapping route - failed to serialize request body")
	}
	r := c.NewRequestWithBody("POST", requestURL, buf)
	resp, err := c.DoRequest(r)
	if err != nil {
		return mr, errors.Wrap(err, "Error creating route")
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return mr, errors.Wrap(err, "Error creating route")
	}
	err = json.Unmarshal(resBody, &mr)
	if err != nil {
		return mr, errors.Wrap(err, "Error unmarshalling routes")
	}
	return mr, nil

}
func (c *Client) ListRoutesByQuery(query url.Values) ([]Route, error) {
	return c.fetchRoutes("/v2/routes?" + query.Encode())
}

func (c *Client) fetchRoutes(requestUrl string) ([]Route, error) {
	var routes []Route
	for {
		routesResp, err := c.getRoutesResponse(requestUrl)
		if err != nil {
			return []Route{}, err
		}
		for _, route := range routesResp.Resources {
			route.Entity.Guid = route.Meta.GUID
			route.Entity.c = c
			routes = append(routes, route.Entity)
		}
		requestUrl = routesResp.NextUrl
		if requestUrl == "" {
			break
		}
	}
	return routes, nil
}

func (c *Client) ListRoutes() ([]Route, error) {
	return c.ListRoutesByQuery(nil)
}

func (c *Client) getRoutesResponse(requestUrl string) (RoutesResponse, error) {
	var routesResp RoutesResponse
	r := c.NewRequest("GET", requestUrl)
	resp, err := c.DoRequest(r)
	if err != nil {
		return RoutesResponse{}, errors.Wrap(err, "Error requesting routes")
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return RoutesResponse{}, errors.Wrap(err, "Error reading routes body")
	}
	err = json.Unmarshal(resBody, &routesResp)
	if err != nil {
		return RoutesResponse{}, errors.Wrap(err, "Error unmarshalling routes")
	}
	return routesResp, nil
}

func (c *Client) createRoute(requestUrl string, routeRequest RouteRequest) (RoutesResource, error) {
	var routeResp RoutesResource
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(routeRequest)
	if err != nil {
		return RoutesResource{}, errors.Wrap(err, "Error creating route - failed to serialize request body")
	}
	r := c.NewRequestWithBody("POST", requestUrl, buf)
	resp, err := c.DoRequest(r)
	if err != nil {
		return RoutesResource{}, errors.Wrap(err, "Error creating route")
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return RoutesResource{}, errors.Wrap(err, "Error creating route")
	}
	err = json.Unmarshal(resBody, &routeResp)
	if err != nil {
		return RoutesResource{}, errors.Wrap(err, "Error unmarshalling routes")
	}
	return routeResp, nil
}
