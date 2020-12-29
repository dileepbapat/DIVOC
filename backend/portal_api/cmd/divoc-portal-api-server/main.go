package main

import (
	"github.com/divoc/portal-api/config"
	"github.com/divoc/portal-api/pkg"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/divoc/portal-api/swagger_gen/restapi"
	"github.com/divoc/portal-api/swagger_gen/restapi/operations"
	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	err := configor.Load(&config.Config, "./config/application-default.yml",
		//"config/application.yml"
	)
	if err != nil {
		panic("Unable to read configurations")
	}
	pkg.InitClickHouseConnection()

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewDivocPortalAPIAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Divoc Portal API"
	parser.LongDescription = "Digital infra for vaccination certificates"
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	server.ConfigureAPI()

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
