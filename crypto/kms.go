package utilsCrypto

import (
	"context"
	b64 "encoding/base64"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

type AgEncryption struct {
	kmsClient *kms.Client
}

func BuildAgEncryption() (*AgEncryption, error) {
	kmsClient, err := generateKmsClient()
	if err != nil {
		return nil, err
	}
	return &AgEncryption{kmsClient: kmsClient}, nil
}

func (a *AgEncryption) EncryptText(ctx context.Context, keyId, textForEncryption string) (string, error) {
	input := &kms.EncryptInput{
		KeyId:               &keyId,
		Plaintext:           []byte(textForEncryption),
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
	}

	outputEncryption, err := a.kmsClient.Encrypt(ctx, input)
	if err != nil {
		return "", err
	}

	blobString := b64.StdEncoding.EncodeToString(outputEncryption.CiphertextBlob)
	return blobString, nil
}

func (a *AgEncryption) DecryptText(ctx context.Context, keyId, encryptedText string) (string, error) {
	blob, err := b64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}
	input := &kms.DecryptInput{
		KeyId:               &keyId,
		CiphertextBlob:      blob,
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
	}
	outputDecryption, err := a.kmsClient.Decrypt(ctx, input)
	if err != nil {
		return "", err
	}
	return string(outputDecryption.Plaintext), nil
}

func generateKmsClient() (*kms.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := kms.NewFromConfig(cfg)
	return client, nil
}
