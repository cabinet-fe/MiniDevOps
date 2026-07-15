package service

// CredentialSecretResolver adapts CredentialService for engine.SecretResolver.
type CredentialSecretResolver struct {
	creds *CredentialService
}

func NewCredentialSecretResolver(creds *CredentialService) *CredentialSecretResolver {
	return &CredentialSecretResolver{creds: creds}
}

func (r *CredentialSecretResolver) Resolve(id uint) (typ, username, secret, passphrase string, err error) {
	c, secret, passphrase, err := r.creds.GetDecrypted(id)
	if err != nil {
		return "", "", "", "", err
	}
	return c.Type, c.Username, secret, passphrase, nil
}
