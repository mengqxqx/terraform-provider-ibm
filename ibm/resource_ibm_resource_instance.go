// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"context"
	//"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gc "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	rc "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	rsInstanceSuccessStatus      = "active"
	rsInstanceProgressStatus     = "in progress"
	rsInstanceProvisioningStatus = "provisioning"
	rsInstanceInactiveStatus     = "inactive"
	rsInstanceFailStatus         = "failed"
	rsInstanceRemovedStatus      = "removed"
	rsInstanceReclamation        = "pending_reclamation"
)

func resourceIBMResourceInstance() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMResourceInstanceCreate,
		Read:     resourceIBMResourceInstanceRead,
		Update:   resourceIBMResourceInstanceUpdate,
		Delete:   resourceIBMResourceInstanceDelete,
		Exists:   resourceIBMResourceInstanceExists,
		Importer: &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				return resourceTagsCustomizeDiff(diff)
			},
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for the resource instance",
			},

			"service": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the service offering like cloud-object-storage, kms etc",
			},

			"plan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The plan type of the service",
			},

			"location": {
				Description: "The location where the instance available",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},

			"resource_group_id": {
				Description: "The resource group id",
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
				Computed:    true,
			},

			"parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Arbitrary parameters to pass. Must be a JSON object",
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceIBMVPCHash,
			},

			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of resource instance",
			},

			"crn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CRN of resource instance",
			},

			"guid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Guid of resource instance",
			},

			"service_endpoints": {
				Description:  "Types of the service endpoints. Possible values are 'public', 'private', 'public-and-private'.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAllowedStringValue([]string{"public", "private", "public-and-private"}),
			},

			"dashboard_url": {
				Description: "Dashboard URL to access resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"resource_group_crn": {
				Description: "The long ID (full CRN) of the resource group",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"resource_id": {
				Description: "The unique ID of the offering",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"resource_plan_id": {
				Description: "The unique ID of the plan associated with the offering",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"target_crn": {
				Description: "The full deployment CRN as defined in the global catalog",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"state": {
				Description: "The current state of the instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"allow_cleanup": {
				Description: "A boolean that dictates if the resource instance should be deleted (cleaned up) during the processing of a region instance delete call.",
				Type:        schema.TypeBool,
				Computed:    true,
			},

			"last_operation": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The status of the last operation requested on the instance",
			},

			"resource_aliases_url": {
				Description: "The relative path to the resource aliases for the instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"resource_bindings_url": {
				Description: "The relative path to the resource bindings for the instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"resource_keys_url": {
				Description: "The relative path to the resource keys for the instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: "The date when the instance was created.",
				Computed:    true,
			},

			"created_by": {
				Type:        schema.TypeString,
				Description: "The subject who created the instance.",
				Computed:    true,
			},

			"update_at": {
				Type:        schema.TypeString,
				Description: "The date when the instance was last updated.",
				Computed:    true,
			},

			"update_by": {
				Type:        schema.TypeString,
				Description: "The subject who updated the instance.",
				Computed:    true,
			},

			"deleted_at": {
				Type:        schema.TypeString,
				Description: "The date when the instance was deleted.",
				Computed:    true,
			},

			"deleted_by": {
				Type:        schema.TypeString,
				Description: "The subject who deleted the instance.",
				Computed:    true,
			},

			ResourceName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the resource",
			},

			ResourceCRN: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The crn of the resource",
			},

			ResourceStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the resource",
			},

			ResourceGroupName: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name in which resource is provisioned",
			},
			ResourceControllerURL: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the IBM Cloud dashboard that can be used to explore and view details about the resource",
			},

			"extensions": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The extended metadata as a map associated with the resource instance.",
			},
		},
	}
}

func resourceIBMResourceInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return err
	}

	serviceName := d.Get("service").(string)
	plan := d.Get("plan").(string)
	name := d.Get("name").(string)
	location := d.Get("location").(string)

	service, err := gc.NewGlobalCatalogV1UsingExternalConfig(
		&gc.GlobalCatalogV1Options{},
	)
	if err != nil {
		return fmt.Errorf("Create globalcatalog service error: %s", err)
	}
	listCatalogEntriesOptions := service.NewListCatalogEntriesOptions()
	include := "true"
	listCatalogEntriesOptions.Include = &include
	newServiceName := serviceName + " rc:true"
	listCatalogEntriesOptions.Q = &newServiceName
	catalogEntry, response, err := service.ListCatalogEntries(listCatalogEntriesOptions)
	if err != nil {
		return fmt.Errorf("List catalog entries error: %s", err)
	}

	serviceId := *catalogEntry.Resources[0].ID

	fmt.Println("Resp code:", response.StatusCode)
	fmt.Println("Entry:", *catalogEntry.ResourceCount)
	fmt.Println("Resources:", *catalogEntry.Resources[0].ID)

	//Start to get servicePlanID
	servicePlanID := plan
	getChildObjectsOptions := service.NewGetChildObjectsOptions(
		serviceId,
		"*",
	)

	planResult, response, err := service.GetChildObjects(getChildObjectsOptions)
	if err != nil {
		return fmt.Errorf("GetChildObjects for servicePlanID error: %s", err)
	}

	fmt.Println(len(planResult.Resources))

	resources := planResult.Resources
	for _, v := range resources {
		if *v.Name == plan {
			servicePlanID = *v.ID
			fmt.Println("The plan id:", servicePlanID)
			break
		}
	}

	// Start to get deployment to set catalog_crn
	catalogCRN := ""
	getChildObjectsOptions = service.NewGetChildObjectsOptions(
		servicePlanID,
		"*",
	)

	deploymentResult, response, err := service.GetChildObjects(getChildObjectsOptions)
	if err != nil {
		return fmt.Errorf("GetChildObjects for deployment error: %s", err)
	}

	fmt.Println(len(deploymentResult.Resources))

	resources = deploymentResult.Resources
	for _, v := range resources {
		fmt.Println("HHH", *v.Name, *v.ID, *v.Metadata.Deployment.Location, *v.CatalogCRN)

		if *v.Metadata.Deployment.Location == location {
			catalogCRN = *v.CatalogCRN
			fmt.Println("The catalogCRN:", catalogCRN)
			break
		}
	}

	rsInst := rc.CreateResourceInstanceOptions{
		Name:           &name,
		Target:         &catalogCRN,
		ResourcePlanID: &servicePlanID,
	}

	if rsGrpID, ok := d.GetOk("resource_group_id"); ok {
		rsInst.ResourceGroup = rsGrpID.(*string)
	} else {
		defaultRg, err := defaultResourceGroup(meta)
		if err != nil {
			return err
		}
		rsInst.ResourceGroup = &defaultRg
	}
	if _, ok := d.GetOk("tags"); ok {
		rsInst.SetTags(d.Get("tags").([]string))
	}
	if _, ok := d.GetOk("allow_cleanup"); ok {
		rsInst.SetAllowCleanup(d.Get("allow_cleanup").(bool))
	}
	if _, ok := d.GetOk("entity_lock"); ok {
		rsInst.SetEntityLock(d.Get("entity_lock").(bool))
	}

	params := map[string]interface{}{}
	if serviceEndpoints, ok := d.GetOk("service_endpoints"); ok {
		params["service-endpoints"] = serviceEndpoints.(string)
	}
	if parameters, ok := d.GetOk("parameters"); ok {
		temp := parameters.(map[string]interface{})
		for k, v := range temp {
			if v == "true" || v == "false" {
				b, _ := strconv.ParseBool(v.(string))
				params[k] = b
			} else if strings.HasPrefix(v.(string), "[") && strings.HasSuffix(v.(string), "]") {
				//transform v.(string) to be []string
				arrayString := v.(string)
				trimLeft := strings.TrimLeft(arrayString, "[")
				trimRight := strings.TrimRight(trimLeft, "]")
				array := strings.Split(trimRight, ",")
				result := []string{}
				for _, a := range array {
					result = append(result, strings.Trim(a, "\""))
				}
				params[k] = result
			} else {
				params[k] = v
			}
		}

	}

	rsInst.SetParameters(params)

	instance, resp, err := rsConClient.CreateResourceInstance(&rsInst)
	if err != nil {
		return fmt.Errorf("Error when creating resource instance: %s with resp code: %s, NAME->%s, LOCATION->%s, GROUP_ID->%s, PLAN_ID->%s", err, resp, *rsInst.Name, *rsInst.Target, *rsInst.ResourceGroup, *rsInst.ResourcePlanID)
	}

	d.SetId(*instance.ID)

	_, err = waitForResourceInstanceCreate(d, meta)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for create resource instance (%s) to be succeeded: %s", d.Id(), err)
	}

	v := os.Getenv("IC_ENV_TAGS")
	if _, ok := d.GetOk("tags"); ok || v != "" {
		oldList, newList := d.GetChange("tags")
		err = UpdateTagsUsingCRN(oldList, newList, meta, *instance.CRN)
		if err != nil {
			log.Printf(
				"Error on create of resource instance (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceIBMResourceInstanceRead(d, meta)
}

func resourceIBMResourceInstanceRead(d *schema.ResourceData, meta interface{}) error {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return err
	}

	instanceID := d.Id()
	resourceInstanceGet := rc.GetResourceInstanceOptions{
		ID: &instanceID,
	}

	instance, resp, err := rsConClient.GetResourceInstance(&resourceInstanceGet)
	if err != nil {
		return fmt.Errorf("Error retrieving resource instance: %s with resp code: %s", err, resp)
	}

	tags, err := GetTagsUsingCRN(meta, *instance.CRN)
	if err != nil {
		log.Printf(
			"Error on get of resource instance tags (%s) tags: %s", d.Id(), err)
	}
	d.Set("tags", tags)
	d.Set("name", instance.Name)
	d.Set("status", instance.State)
	d.Set("resource_group_id", instance.ResourceGroupID)
	//d.Set("location", instance.RegionID)
	d.Set("crn", instance.CRN)
	d.Set("dashboard_url", instance.DashboardURL)

	rsCatClient, err := meta.(ClientSession).ResourceCatalogAPI()
	if err != nil {
		return err
	}
	rsCatRepo := rsCatClient.ResourceCatalog()

	serviceOff, err := rsCatRepo.GetServiceName(*instance.ResourceID)
	if err != nil {
		return fmt.Errorf("Error retrieving service offering: %s", err)
	}

	d.Set("service", serviceOff)

	d.Set(ResourceName, instance.Name)
	d.Set(ResourceCRN, instance.CRN)
	d.Set(ResourceStatus, instance.State)
	d.Set(ResourceGroupName, instance.ResourceGroupCRN)

	rcontroller, err := getBaseController(meta)
	if err != nil {
		return err
	}
	d.Set(ResourceControllerURL, rcontroller+"/services/")

	servicePlan, err := rsCatRepo.GetServicePlanName(*instance.ResourcePlanID)
	if err != nil {
		return fmt.Errorf("Error retrieving plan: %s", err)
	}
	d.Set("plan", servicePlan)
	d.Set("guid", instance.GUID)
	if instance.Parameters != nil {
		if endpoint, ok := instance.Parameters["service-endpoints"]; ok {
			d.Set("service_endpoints", endpoint)
		}
	}

	if len(instance.Extensions) == 0 {
		d.Set("extensions", instance.Extensions)
	} else {
		d.Set("extensions", Flatten(instance.Extensions))
	}

	return nil
}

func resourceIBMResourceInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return err
	}

	instanceID := d.Id()

	resourceInstanceUpdate := rc.UpdateResourceInstanceOptions{
		ID: &instanceID,
	}
	if d.HasChange("name") {
		name := d.Get("name").(string)
		resourceInstanceUpdate.Name = &name
	}

	if d.HasChange("plan") {
		plan := d.Get("plan").(string)
		service := d.Get("service").(string)
		rsCatClient, err := meta.(ClientSession).ResourceCatalogAPI()
		if err != nil {
			return err
		}
		rsCatRepo := rsCatClient.ResourceCatalog()

		serviceOff, err := rsCatRepo.FindByName(service, true)
		if err != nil {
			return fmt.Errorf("Error retrieving service offering: %s", err)
		}

		servicePlan, err := rsCatRepo.GetServicePlanID(serviceOff[0], plan)
		if err != nil {
			return fmt.Errorf("Error retrieving plan: %s", err)
		}

		resourceInstanceUpdate.ResourcePlanID = &servicePlan

	}
	params := map[string]interface{}{}

	if d.HasChange("service_endpoints") {
		endpoint := d.Get("service_endpoints").(string)
		params["service-endpoints"] = endpoint
	}

	resourceInstanceGet := rc.GetResourceInstanceOptions{
		ID: &instanceID,
	}
	if d.HasChange("parameters") {
		instance, resp, err := rsConClient.GetResourceInstance(&resourceInstanceGet)
		if err != nil {
			return fmt.Errorf("Error retrieving resource instance: %s with resp code: %s", err, resp)
		}

		if parameters, ok := d.GetOk("parameters"); ok {
			temp := parameters.(map[string]interface{})
			for k, v := range temp {
				if v == "true" || v == "false" {
					b, _ := strconv.ParseBool(v.(string))
					params[k] = b
				} else if strings.HasPrefix(v.(string), "[") && strings.HasSuffix(v.(string), "]") {
					//transform v.(string) to be []string
					arrayString := v.(string)
					trimLeft := strings.TrimLeft(arrayString, "[")
					trimRight := strings.TrimRight(trimLeft, "]")
					array := strings.Split(trimRight, ",")
					result := []string{}
					for _, a := range array {
						result = append(result, strings.Trim(a, "\""))
					}
					params[k] = result
				} else {
					params[k] = v
				}
			}
		}
		serviceEndpoints := d.Get("service_endpoints").(string)
		if serviceEndpoints != "" {
			endpoint := d.Get("service_endpoints").(string)
			params["service-endpoints"] = endpoint
		} else if _, ok := instance.Parameters["service-endpoints"]; ok {
			params["service-endpoints"] = instance.Parameters["service-endpoints"]
		}

	}
	if d.HasChange("service_endpoints") || d.HasChange("parameters") {
		resourceInstanceUpdate.Parameters = params
	}

	instance, resp, err := rsConClient.GetResourceInstance(&resourceInstanceGet)
	if err != nil {
		return fmt.Errorf("Error Getting resource instance: %s with resp code: %s", err, resp)
	}

	if d.HasChange("tags") {
		oldList, newList := d.GetChange(isVPCTags)
		err = UpdateTagsUsingCRN(oldList, newList, meta, *instance.CRN)
		if err != nil {
			log.Printf(
				"Error on update of resource instance (%s) tags: %s", d.Id(), err)
		}
	}

	_, resp, err = rsConClient.UpdateResourceInstance(&resourceInstanceUpdate)
	if err != nil {
		return fmt.Errorf("Error updating resource instance: %s with resp code: %s", err, resp)
	}

	_, err = waitForResourceInstanceUpdate(d, meta)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for update resource instance (%s) to be succeeded: %s", d.Id(), err)
	}

	return resourceIBMResourceInstanceRead(d, meta)
}

func resourceIBMResourceInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return err
	}
	id := d.Id()
	resourceInstanceDelete := rc.DeleteResourceInstanceOptions{
		ID: &id,
	}

	resp, error := rsConClient.DeleteResourceInstance(&resourceInstanceDelete)
	if error != nil {
		return fmt.Errorf("Error deleting resource instance: %s with resp code: %s", error, resp)
	}

	_, err = waitForResourceInstanceDelete(d, meta)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for resource instance (%s) to be deleted: %s", d.Id(), err)
	}

	d.SetId("")

	return nil
}
func resourceIBMResourceInstanceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return false, err
	}
	instanceID := d.Id()
	resourceInstanceGet := rc.GetResourceInstanceOptions{
		ID: &instanceID,
	}

	instance, resp, err := rsConClient.GetResourceInstance(&resourceInstanceGet)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error communicating with the API: %s with resp code: %s", err, resp)
	}

	return *instance.ID == instanceID, nil
}

func waitForResourceInstanceCreate(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return false, err
	}
	instanceID := d.Id()
	resourceInstanceGet := rc.GetResourceInstanceOptions{
		ID: &instanceID,
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{rsInstanceProgressStatus, rsInstanceInactiveStatus, rsInstanceProvisioningStatus},
		Target:  []string{rsInstanceSuccessStatus},
		Refresh: func() (interface{}, string, error) {
			instance, resp, err := rsConClient.GetResourceInstance(&resourceInstanceGet)
			if err != nil {
				if resp != nil && resp.StatusCode == 404 {
					return nil, "", fmt.Errorf("The resource instance %s does not exist anymore: %v", d.Id(), err)
				}
				return nil, "", fmt.Errorf("Get the resource instance %s failed with resp code: %s, err: %v", d.Id(), resp, err)
			}
			if *instance.State == rsInstanceFailStatus {
				return instance, *instance.State, fmt.Errorf("The resource instance %s failed: %v", d.Id(), err)
			}
			return instance, *instance.State, nil
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func waitForResourceInstanceUpdate(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return false, err
	}
	instanceID := d.Id()
	resourceInstanceGet := rc.GetResourceInstanceOptions{
		ID: &instanceID,
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{rsInstanceProgressStatus, rsInstanceInactiveStatus},
		Target:  []string{rsInstanceSuccessStatus},
		Refresh: func() (interface{}, string, error) {
			instance, resp, err := rsConClient.GetResourceInstance(&resourceInstanceGet)
			if err != nil {
				if resp != nil && resp.StatusCode == 404 {
					return nil, "", fmt.Errorf("The resource instance %s does not exist anymore: %v", d.Id(), err)
				}
				return nil, "", fmt.Errorf("Get the resource instance %s failed with resp code: %s, err: %v", d.Id(), resp, err)
			}
			if *instance.State == rsInstanceFailStatus {
				return instance, *instance.State, fmt.Errorf("The resource instance %s failed: %v", d.Id(), err)
			}
			return instance, *instance.State, nil
		},
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func waitForResourceInstanceDelete(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	rsConClient, err := meta.(ClientSession).ResourceControllerV2API()
	if err != nil {
		return false, err
	}
	instanceID := d.Id()
	resourceInstanceGet := rc.GetResourceInstanceOptions{
		ID: &instanceID,
	}
	stateConf := &resource.StateChangeConf{
		Pending: []string{rsInstanceProgressStatus, rsInstanceInactiveStatus, rsInstanceSuccessStatus},
		Target:  []string{rsInstanceRemovedStatus, rsInstanceReclamation},
		Refresh: func() (interface{}, string, error) {
			instance, resp, err := rsConClient.GetResourceInstance(&resourceInstanceGet)
			if err != nil {
				if resp != nil && resp.StatusCode == 404 {
					return instance, rsInstanceSuccessStatus, nil
				}
				return nil, "", fmt.Errorf("Get the resource instance %s failed with resp code: %s, err: %v", d.Id(), resp, err)
			}
			if *instance.State == rsInstanceFailStatus {
				return instance, *instance.State, fmt.Errorf("The resource instance %s failed to delete: %v", d.Id(), err)
			}
			return instance, *instance.State, nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}
