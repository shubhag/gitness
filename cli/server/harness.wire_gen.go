// Code generated by Wire. DO NOT EDIT.

//go:build !wireinject && harness
// +build !wireinject,harness

package server

import (
	"context"

	"github.com/harness/gitness/gitrpc"
	server2 "github.com/harness/gitness/gitrpc/server"
	"github.com/harness/gitness/harness/auth/authn"
	"github.com/harness/gitness/harness/auth/authz"
	"github.com/harness/gitness/harness/bootstrap"
	"github.com/harness/gitness/harness/client"
	"github.com/harness/gitness/harness/router"
	"github.com/harness/gitness/harness/store"
	types2 "github.com/harness/gitness/harness/types"
	"github.com/harness/gitness/harness/types/check"
	"github.com/harness/gitness/internal/api/controller/pullreq"
	"github.com/harness/gitness/internal/api/controller/repo"
	"github.com/harness/gitness/internal/api/controller/service"
	"github.com/harness/gitness/internal/api/controller/serviceaccount"
	"github.com/harness/gitness/internal/api/controller/space"
	"github.com/harness/gitness/internal/api/controller/user"
	"github.com/harness/gitness/internal/cron"
	router2 "github.com/harness/gitness/internal/router"
	"github.com/harness/gitness/internal/server"
	"github.com/harness/gitness/internal/store/database"
	"github.com/harness/gitness/types"
)

// Injectors from harness.wire.go:

func initSystem(ctx context.Context, config *types.Config) (*system, error) {
	checkUser := check.ProvideUserCheck()
	typesConfig, err := types2.LoadConfig()
	if err != nil {
		return nil, err
	}
	serviceJWTProvider, err := client.ProvideServiceJWTProvider(typesConfig)
	if err != nil {
		return nil, err
	}
	aclClient, err := client.ProvideACLClient(serviceJWTProvider, typesConfig)
	if err != nil {
		return nil, err
	}
	authorizer := authz.ProvideAuthorizer(typesConfig, aclClient)
	db, err := database.ProvideDatabase(ctx, config)
	if err != nil {
		return nil, err
	}
	principalUIDTransformation := store.ProvidePrincipalUIDTransformation()
	userStore := database.ProvideUserStore(db, principalUIDTransformation)
	tokenStore := database.ProvideTokenStore(db)
	controller := user.NewController(checkUser, authorizer, userStore, tokenStore)
	checkService := check.ProvideServiceCheck()
	serviceStore := database.ProvideServiceStore(db, principalUIDTransformation)
	serviceController := service.NewController(checkService, authorizer, serviceStore)
	bootstrapBootstrap := bootstrap.ProvideBootstrap(config, controller, serviceController)
	tokenClient, err := client.ProvideTokenClient(serviceJWTProvider, typesConfig)
	if err != nil {
		return nil, err
	}
	userClient, err := client.ProvideUserClient(serviceJWTProvider, typesConfig)
	if err != nil {
		return nil, err
	}
	serviceAccountClient, err := client.ProvideServiceAccountClient(serviceJWTProvider, typesConfig)
	if err != nil {
		return nil, err
	}
	serviceAccount := check.ProvideServiceAccountCheck()
	serviceAccountStore := database.ProvideServiceAccountStore(db, principalUIDTransformation)
	pathTransformation := store.ProvidePathTransformation()
	spaceStore := database.ProvideSpaceStore(db, pathTransformation)
	repoStore := database.ProvideRepoStore(db, pathTransformation)
	serviceaccountController := serviceaccount.NewController(serviceAccount, authorizer, serviceAccountStore, spaceStore, repoStore, tokenStore)
	checkSpace := check.ProvideSpaceCheck()
	spaceController := space.ProvideController(config, checkSpace, authorizer, spaceStore, repoStore, serviceAccountStore)
	accountClient, err := client.ProvideAccountClient(serviceJWTProvider, typesConfig)
	if err != nil {
		return nil, err
	}
	authenticator, err := authn.ProvideAuthenticator(controller, tokenClient, userClient, typesConfig, serviceAccountClient, serviceaccountController, serviceController, spaceController, accountClient)
	if err != nil {
		return nil, err
	}
	checkRepo := check.ProvideRepoCheck()
	gitrpcConfig := ProvideGitRPCClientConfig(config)
	gitrpcInterface, err := gitrpc.ProvideClient(gitrpcConfig)
	if err != nil {
		return nil, err
	}
	repoController := repo.ProvideController(config, checkRepo, authorizer, spaceStore, repoStore, serviceAccountStore, gitrpcInterface)
	pullReqStore := database.ProvidePullReqStore(db)
	pullreqController := pullreq.ProvideController(db, authorizer, pullReqStore, repoStore, serviceAccountStore, gitrpcInterface)
	apiHandler := router.ProvideAPIHandler(config, authenticator, accountClient, spaceController, repoController, pullreqController)
	gitHandler := router.ProvideGitHandler(config, repoStore, authenticator, authorizer, gitrpcInterface)
	webHandler := router2.ProvideWebHandler(config)
	routerRouter := router2.ProvideRouter(apiHandler, gitHandler, webHandler)
	serverServer := server.ProvideServer(config, routerRouter)
	serverConfig := ProvideGitRPCServerConfig(config)
	server3, err := server2.ProvideServer(serverConfig)
	if err != nil {
		return nil, err
	}
	nightly := cron.NewNightly()
	serverSystem := newSystem(bootstrapBootstrap, serverServer, server3, nightly)
	return serverSystem, nil
}
