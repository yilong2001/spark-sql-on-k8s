package controller

import (
	"context"
	"fmt"
	//"encoding/json" 
	//"os"
	"net/url"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func (r *SparkAppCtrl) doUploadToS3(s3ns, s3path string, 
	srcfiles []string) ([]string, error) {
	uploadLocationUrl, err := url.Parse(r.cfg.S3UploadDir)
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	uploadBucket := uploadLocationUrl.Host
	log.Infof("upload file to s3 : %s, %s, %s \n", uploadLocationUrl, uploadLocationUrl.Scheme, uploadBucket)

	var uh *uploadHandler
	ctx := context.Background()
	switch uploadLocationUrl.Scheme {
	case "s3":
		uh, err = newPrivateS3Blob(ctx, uploadBucket, 
			r.cfg.S3Endpoint, r.cfg.S3AccessKey, r.cfg.S3SecretKey, "")
	default:
		err = fmt.Errorf("unsupported upload location URL scheme: %s", uploadLocationUrl.Scheme)
		log.Errorf("%v", err)
    	return nil, err
	}

	// Check if bucket has been successfully setup
	if err != nil {
		log.Errorf("%v", err)
    	return nil, err
	}

	uploadPath := filepath.Join(defaultSparkUploadRoot, s3ns, s3path)
	
	outfiles := make([]string, 0)
	for _, srcfile := range srcfiles {
		uploadFilePath, err := uh.uploadToBucket(uploadPath, srcfile, true)
		if err != nil {
			log.Errorf("error : upload local file to s3 failed : %v !", err)
			return nil, err
		}
		outfiles = append(outfiles, uploadFilePath)
	}

	return outfiles, nil
}

