package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func requireDocker(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("requires Docker")
	}
	if !DockerAvailable() {
		t.Skip("Docker not available")
	}
}

func TestMongoDBImporterIntegration(t *testing.T) {
	requireDocker(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongo: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	_, err = client.Database("kiwi_test").Collection("widgets").InsertOne(ctx, bson.M{
		"name":   "alpha",
		"qty":    10,
		"active": true,
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	src, err := NewMongoDB(uri, "kiwi_test", "widgets")
	if err != nil {
		t.Fatalf("NewMongoDB: %v", err)
	}
	defer src.Close()

	tables, err := BrowseMongoCollections(ctx, src)
	if err != nil {
		t.Fatalf("BrowseMongoCollections: %v", err)
	}
	found := false
	for _, tbl := range tables {
		if tbl.Name == "widgets" {
			found = true
		}
	}
	if !found {
		t.Fatalf("widgets collection not listed: %+v", tables)
	}

	records, errs := src.Stream(ctx)
	var got []Record
	for rec := range records {
		got = append(got, rec)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream: %v", err)
		}
	}
	if len(got) != 1 || got[0].Fields["name"] != "alpha" {
		t.Fatalf("records: %+v", got)
	}
}

func TestRedisImporterIntegration(t *testing.T) {
	requireDocker(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := redis.Run(ctx, "redis:7")
	if err != nil {
		t.Fatalf("start redis: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}

	seed := goredis.NewClient(&goredis.Options{Addr: endpoint})
	if err := seed.HSet(ctx, "widget:1", "name", "alpha", "qty", "10").Err(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	seed.Close()

	src, err := NewRedis(endpoint, "", 0, "widget:*")
	if err != nil {
		t.Fatalf("NewRedis: %v", err)
	}
	defer src.Close()

	records, errs := src.Stream(ctx)
	var got []Record
	for rec := range records {
		got = append(got, rec)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream: %v", err)
		}
	}
	if len(got) != 1 || got[0].Fields["name"] != "alpha" {
		t.Fatalf("records: %+v", got)
	}
}

func TestElasticsearchImporterIntegration(t *testing.T) {
	requireDocker(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	container, err := elasticsearch.Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:8.11.0",
		elasticsearch.WithPassword("changeme"),
	)
	if err != nil {
		t.Fatalf("start elasticsearch: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	mapped, err := container.MappedPort(ctx, "9200/tcp")
	if err != nil {
		t.Fatalf("port: %v", err)
	}
	base := fmt.Sprintf("http://elastic:changeme@%s:%s", host, mapped.Port())
	doc := map[string]any{"name": "alpha", "qty": 10}
	body, _ := json.Marshal(doc)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, base+"/widgets/_doc/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("index doc: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("index status: %d", resp.StatusCode)
	}
	refresh, _ := http.NewRequestWithContext(ctx, http.MethodPost, base+"/widgets/_refresh", nil)
	if rresp, err := http.DefaultClient.Do(refresh); err == nil {
		rresp.Body.Close()
	}

	src, err := NewElasticsearch(base, "widgets", nil)
	if err != nil {
		t.Fatalf("NewElasticsearch: %v", err)
	}

	records, errs := src.Stream(ctx)
	var got []Record
	for rec := range records {
		got = append(got, rec)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream: %v", err)
		}
	}
	if len(got) != 1 {
		t.Fatalf("records=%d, want 1: %+v", len(got), got)
	}
}

func TestDynamoDBImporterIntegration(t *testing.T) {
	requireDocker(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForListeningPort("8000/tcp").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start dynamodb-local: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	port, err := container.MappedPort(ctx, "8000/tcp")
	if err != nil {
		t.Fatalf("port: %v", err)
	}
	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_ENDPOINT_URL", endpoint)

	ddbClient := dynamodb.NewFromConfig(aws.Config{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(endpoint),
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"}, nil
		}),
	})
	_, err = ddbClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("widgets"),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	_, err = ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("widgets"),
		Item: map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: "1"},
			"name": &types.AttributeValueMemberS{Value: "alpha"},
			"qty":  &types.AttributeValueMemberN{Value: "10"},
		},
	})
	if err != nil {
		t.Fatalf("put item: %v", err)
	}

	src, err := NewDynamoDB("us-east-1", "widgets")
	if err != nil {
		t.Fatalf("NewDynamoDB: %v", err)
	}

	records, errs := src.Stream(ctx)
	var got []Record
	for rec := range records {
		got = append(got, rec)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream: %v", err)
		}
	}
	if len(got) != 1 || got[0].PrimaryKey == "" {
		t.Fatalf("records: %+v", got)
	}
}
