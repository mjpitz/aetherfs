// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package s3

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"

	blockv1 "github.com/mjpitz/aetherfs/api/aetherfs/block/v1"
	datasetv1 "github.com/mjpitz/aetherfs/api/aetherfs/dataset/v1"
	"github.com/mjpitz/myago/livetls"
)

type Config struct {
	Endpoint        string         `json:"endpoint"          usage:"location of s3 endpoint" default:"s3.amazonaws.com"`
	TLS             livetls.Config `json:"tls"`
	AccessKeyID     string         `json:"access_key_id"     usage:"the access key id used to identify the client"`
	SecretAccessKey string         `json:"secret_access_key" usage:"the secret access key used to authenticate the client"`
	Region          string         `json:"region"            usage:"the region where the bucket exists"`
	Bucket          string         `json:"bucket"            usage:"the name of the bucket to use"`
}

func ObtainStores(ctx context.Context, cfg Config) (blockv1.BlockAPIServer, datasetv1.DatasetAPIServer, error) {
	tls, err := livetls.New(ctx, cfg.TLS)
	if err != nil {
		return nil, nil, err
	}

	var rt http.RoundTripper
	if tls != nil {
		rt = &http.Transport{
			TLSClientConfig: tls,
			// pulled from http.DefaultTransport
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	// support common AWS and MinIO environment variables
	providers := []credentials.Provider{
		&credentials.EnvAWS{},
		&credentials.EnvMinio{},
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		providers = append(providers, &credentials.Static{
			Value: credentials.Value{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
				SignerType:      credentials.SignatureV4,
			},
		})
	}

	s3Client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewChainCredentials(providers),
		Secure:    tls != nil,
		Transport: rt,
		Region:    cfg.Region,
	})

	if err != nil {
		return nil, nil, err
	}

	err = s3Client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
	exists, _ := s3Client.BucketExists(ctx, cfg.Bucket)
	if !exists {
		return nil, nil, errors.Wrap(err, "bucket does not exist")
	}

	blockSvc := &blockService{
		s3Client:   s3Client,
		bucketName: cfg.Bucket,
	}

	datasetSvc := &datasetService{
		s3Client:   s3Client,
		bucketName: cfg.Bucket,
	}

	return blockSvc, datasetSvc, nil
}
