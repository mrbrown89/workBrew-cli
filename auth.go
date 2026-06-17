package main

import "github.com/zalando/go-keyring"

const keychainService = "workbrew-cli"
const keychainAccount = "default"

func saveAPIToken(token string) error {
	return keyring.Set(keychainService, keychainAccount, token)
}

func loadAPIToken() (string, error) {
	return keyring.Get(keychainService, keychainAccount)
}
