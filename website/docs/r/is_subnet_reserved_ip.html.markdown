---
subcategory: "VPC infrastructure"
layout: "ibm"
page_title: "IBM : ibm_is_subnet_reserved_ip"
description: |-
  Manages IBM Subnet reserved IP
---

# ibm_is_virtual_endpoint_gateway

Provides a subnet reserved IP resource. This allows Subnet reserved IP to be created, updated, and deleted.

## Example Usage

In the following example, you can create a Reserved IP:

```hcl
    // Create a VPC
    resource "ibm_is_vpc" "vpc1" {
        name = "my-vpc"
    }

    // Create a subnet
    resource "ibm_is_subnet" "subnet1" {
        name                     = "my-subnet"
        vpc                      = ibm_is_vpc.vpc1.id
        zone                     = "us-south-1"
        total_ipv4_address_count = 256
    }

    // Create the resrved IP in the following ways

    // Only with Subnet ID
    resource "ibm_is_subnet_reserved_ip" "res_ip" {
        subnet = ibm_is_subnet.subnet1.id
    }

    // Subnet ID with a given name
    resource "ibm_is_subnet_reserved_ip" "res_ip_name" {
        subnet = ibm_is_subnet.subnet1.id
        name = "my-subnet"
    }

    // Subnet ID with auto_delete
    resource "ibm_is_subnet_reserved_ip" "res_ip_auto_delete" {
        subnet = ibm_is_subnet.subnet1.id
        auto_delete = true
    }

    // Subnet ID with both name and auto_delete
    resource "ibm_is_subnet_reserved_ip" "res_ip_auto_delete_name" {
        subnet = ibm_is_subnet.subnet1.id
        name = "my-subnet"
        auto_delete = true
    }
```

## Argument Reference

The following arguments are supported:

* `subnet` - (Required, Forces new resource, string) The subnet id for the reserved IP.
* `name` - (Optional, string) The name of the reserved IP.
    **NOTE**: Raise error if name is given with a prefix `ibm-`.
* `auto_delete` - (Optional, boolean) If reserved IP is auto deleted.


## Attribure Reference

* `id` - The combination of the subnet ID and reserved IP ID seperated by '/'.
* `reserved_ip` - This refers to only the reserved IP.

## Import

ibm_is_subnet_reserved_ip can be imported using subnet ID and reserved IP ID seperated by '/' eg

```hcl
terraform import ibm_is_subnet_reserved_ip.example 0716-13315ad8-d355-4041-bb60-67430d393607/0716-617de4d8-5e2f-4d4a-b0d6-d221bc230c04
```