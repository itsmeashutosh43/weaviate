//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2023 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package moduleshelper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/require"
	"github.com/weaviate/weaviate/test/helper"
	graphqlhelper "github.com/weaviate/weaviate/test/helper/graphql"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func GetClassCount(t *testing.T, className string, tenantKey string) int64 {
	query := fmt.Sprintf("{Aggregate{%s", className)
	if tenantKey != "" {
		query += fmt.Sprintf("(tenantKey:%q)", tenantKey)
	}
	query += " { meta { count}}}}"
	resp := graphqlhelper.AssertGraphQL(t, helper.RootAuth, query)

	class := resp.Get("Aggregate", className).Result.([]interface{})
	require.Len(t, class, 1)

	meta := class[0].(map[string]interface{})["meta"].(map[string]interface{})

	countPayload := meta["count"].(json.Number)

	count, err := countPayload.Int64()
	require.Nil(t, err)

	return count
}

func CreateTestFiles(t *testing.T, dirPath string) []string {
	count := 5
	filePaths := make([]string, count)
	var fileName string

	for i := 0; i < count; i += 1 {
		fileName = fmt.Sprintf("file_%d.db", i)
		filePaths[i] = filepath.Join(dirPath, fileName)
		file, err := os.Create(filePaths[i])
		if err != nil {
			t.Fatalf("failed to create test file '%s': %s", fileName, err)
		}
		fmt.Fprintf(file, "This is content of db file named %s", fileName)
		file.Close()
	}
	return filePaths
}

func CreateGCSBucket(ctx context.Context, t *testing.T, projectID, bucketName string) {
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	require.Nil(t, err)

	err = client.Bucket(bucketName).Create(ctx, projectID, nil)
	gcsErr, ok := err.(*googleapi.Error)
	if ok {
		// the bucket persists from the previous test.
		// if the bucket already exists, we can proceed
		if gcsErr.Code == http.StatusConflict {
			return
		}
	}
	require.Nil(t, err)
}

func CreateAzureContainer(ctx context.Context, t *testing.T, endpoint, containerName string) {
	connectionString := "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://%s/devstoreaccount1;"
	client, err := azblob.NewClientFromConnectionString(fmt.Sprintf(connectionString, endpoint), nil)
	require.Nil(t, err)

	_, err = client.CreateContainer(ctx, containerName, nil)
	require.Nil(t, err)
}

func DeleteAzureContainer(ctx context.Context, t *testing.T, endpoint, containerName string) {
	connectionString := "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://%s/devstoreaccount1;"
	client, err := azblob.NewClientFromConnectionString(fmt.Sprintf(connectionString, endpoint), nil)
	require.Nil(t, err)

	_, err = client.DeleteContainer(ctx, containerName, nil)
	require.Nil(t, err)
}
