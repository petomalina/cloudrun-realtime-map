# presence service
build/presence:
	pack build -e GOOGLE_BUILDABLE=./presence --builder gcr.io/buildpacks/builder:v1 --publish europe-west3-docker.pkg.dev/petermalina/cloudrun-map/presence

deploy/presence: build/presence
	gcloud run deploy presence \
		--region=europe-west3 \
		--image=europe-west3-docker.pkg.dev/petermalina/cloudrun-map/presence \
		--platform managed \
		--allow-unauthenticated

stage/presence: build/presence
	gcloud beta run deploy presence \
		--region=europe-west3 \
		--image=europe-west3-docker.pkg.dev/petermalina/cloudrun-map/presence \
		--platform managed \
		--allow-unauthenticated \
		--no-traffic \
		--tag beta

# iploc service
build/iploc:
	pack build -e GOOGLE_BUILDABLE=./iploc --builder gcr.io/buildpacks/builder:v1 --publish europe-west3-docker.pkg.dev/petermalina/cloudrun-map/iploc

deploy/iploc: build/iploc
	gcloud run deploy iploc \
		--region=europe-west3 \
		--image=europe-west3-docker.pkg.dev/petermalina/cloudrun-map/iploc \
		--platform managed \
		--no-allow-unauthenticated \
		--set-env-vars=IPSTACK_KEY=efcfb40898a5103f8afaf71588a4d28f

stage/iploc: build/iploc
	gcloud beta run deploy iploc \
		--region=europe-west3 \
		--image=europe-west3-docker.pkg.dev/petermalina/cloudrun-map/iploc \
		--platform managed \
		--no-traffic \
		--no-allow-unauthenticated \
		--set-env-vars=IPSTACK_KEY=efcfb40898a5103f8afaf71588a4d28f \
		--tag beta

# cleanup service
build/cleanup:
	pack build -e GOOGLE_BUILDABLE=./cleanup --builder gcr.io/buildpacks/builder:v1 --publish europe-west3-docker.pkg.dev/petermalina/cloudrun-map/cleanup

deploy/cleanup: build/cleanup
	gcloud run deploy cleanup \
		--region=europe-west3 \
		--image=europe-west3-docker.pkg.dev/petermalina/cloudrun-map/cleanup \
		--platform managed \
		--allow-unauthenticated

stage/cleanup: build/cleanup
	gcloud beta run deploy cleanup \
		--region=europe-west3 \
		--image=europe-west3-docker.pkg.dev/petermalina/cloudrun-map/cleanup \
		--platform managed \
		--no-traffic \
		--allow-unauthenticated \
		--tag beta
