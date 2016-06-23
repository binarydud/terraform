package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewaySwaggerAPI() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewaySwaggerAPICreate,
		Read:   resourceAwsApiGatewaySwaggerAPIRead,
		Delete: resourceAwsApiGatewaySwaggerAPIDelete,
		Update: resourceAwsApiGatewaySwaggerAPIUpdate,

		Schema: map[string]*schema.Schema{
			"swagger": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"failonwarnings": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},

			"updatemode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  "overwrite",
			},
		},
	}
}

func resourceAwsApiGatewaySwaggerAPICreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	swagger := d.Get("swagger").(string)

	req := &apigateway.ImportRestApiInput{
		Body: []byte(swagger),
	}

	if d.Get("failonwarnings") != nil {
		req.FailOnWarnings = aws.Bool(d.Get("failonwarnings").(bool))
	}

	res, err := conn.ImportRestApi(req)

	if err != nil {
		return err
	}

	for w := range res.Warnings {
		log.Printf("[WARN] Swagger import warning: %s", w)
	}
	d.SetId(*res.Id)

	return resourceAwsApiGatewaySwaggerAPIRead(d, meta)
}

func resourceAwsApiGatewaySwaggerAPIUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	swagger := d.Get("swagger").(string)
	mode := d.Get("updatemode").(string)
	id := aws.String(d.Id())
	req := &apigateway.PutRestApiInput{
		Body:      []byte(swagger),
		Mode:      &mode,
		RestApiId: id,
	}
	if d.Get("failonwarnings") != nil {
		req.FailOnWarnings = aws.Bool(d.Get("failonwarnings").(bool))
	}
	res, err := conn.PutRestApi(req)
	if err != nil {
		return err
	}

	for w := range res.Warnings {
		log.Printf("[WARN] Swagger import warning: %s", w)
	}
	d.SetId(*res.Id)
	return resourceAwsApiGatewaySwaggerAPIRead(d, meta)
}

func resourceAwsApiGatewaySwaggerAPIRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	_, err := conn.GetRestApi(&apigateway.GetRestApiInput{
		RestApiId: aws.String(d.Id()),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			d.SetId("")
			return nil
		}
		return err
	}

	return nil
}

func resourceAwsApiGatewaySwaggerAPIDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteRestApi(&apigateway.DeleteRestApiInput{
			RestApiId: aws.String(d.Id()),
		})
		if err == nil {
			return nil
		}

		if apigatewayErr, ok := err.(awserr.Error); ok && apigatewayErr.Code() == "NotFoundException" {
			return nil
		}

		return resource.NonRetryableError(err)
	})
}
