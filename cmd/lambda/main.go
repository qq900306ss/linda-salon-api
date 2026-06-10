// Lambda entrypoint. Lambda Function URLs use the API Gateway v2 payload
// format, so the V2 Gin adapter is used.
package main

import (
	"context"
	"log"

	// Embed the tz database so Asia/Taipei resolves inside the Lambda runtime.
	_ "time/tzdata"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"

	"github.com/qq900306ss/linda-salon-api/internal/app"
)

var ginLambda *ginadapter.GinLambdaV2

func init() {
	router, err := app.Initialize(context.Background())
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	ginLambda = ginadapter.NewV2(router)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
