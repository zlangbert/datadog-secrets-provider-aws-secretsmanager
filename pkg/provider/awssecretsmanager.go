package provider

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
	"github.com/zlangbert/datadog-secrets-provider-aws-secretsmanager/pkg/config"
	"github.com/zlangbert/datadog-secrets-provider-aws-secretsmanager/pkg/secret"
)

type AwsSecretsManagerProvider struct {
	manager *secretsmanager.SecretsManager
}

func NewAwsSecretsManagerProvider(config *config.HelperConfig) (provider SecretProvider, err error) {

	// build aws session
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AWS session")
	}

	provider = &AwsSecretsManagerProvider{
		manager: secretsmanager.New(sess),
	}

	return provider, nil
}

func (provider *AwsSecretsManagerProvider) Resolve(handles []*secret.Handle) (results map[string]secret.Result) {

	// TODO: If multiple keys are desired under one secret, that secret is retrieved multiple times. Optimize by
	//  getting each secret only once
	// evaluate each handle
	secretResults := map[string]secret.Result{}
	for _, handle := range handles {

		// get secret from secrets manager
		s, err := provider.manager.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId: &handle.Id,
		})
		if err != nil {
			secretResults[handle.Handle] = secret.Result{
				Error: fmt.Sprintf("error getting secret: %v", err),
			}
			continue
		}

		// get secret payload
		var secretPayload map[string]string
		err = json.Unmarshal([]byte(*s.SecretString), &secretPayload)
		if err != nil {
			secretResults[handle.Handle] = secret.Result{
				Error: fmt.Sprintf("error parsing secret value, expected object with string key/values: %v", err),
			}
			continue
		}

		// get requested key from secret data
		value := secretPayload[handle.Key]
		if value == "" {
			secretResults[handle.Handle] = secret.Result{
				Error: fmt.Sprintf("secret value for key '%s' is missing or blank", handle.Key),
			}
			continue
		}

		secretResults[handle.Handle] = secret.Result{
			Value: value,
		}
	}

	return secretResults
}
