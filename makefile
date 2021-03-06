CLOUD_RUN_SERVICE_NAME:="go-sample-app"
IMAGE_HOST:=""
IMAGE_NAME:=$(CLOUD_RUN_SERVICE_NAME)
REGION:="asia-northeast1"
TAG:=$(shell date "+%Y%m%d%H%M%S")
MAX_INSTANCE:=2

run-local:
	docker build -t $(IMAGE_NAME):$(TAG) . && \
	docker run --rm -d -p 8080:8080 $(IMAGE_NAME):$(TAG)

run-deploy:
	docker build -t $(IMAGE_NAME):$(TAG) . && \
	docker tag $(IMAGE_NAME):$(TAG) $(IMAGE_HOST)/$(IMAGE_NAME):$(TAG) && \
	docker push $(IMAGE_HOST)/$(IMAGE_NAME):$(TAG)
	gcloud run deploy $(CLOUD_RUN_SERVICE_NAME) \
		--image $(IMAGE_HOST)/$(IMAGE_NAME):$(TAG) \
		--region $(REGION) \
		--platform managed \
		--max-instances $(MAX_INSTANCE)

run-describe:
	gcloud run services describe $(CLOUD_RUN_SERVICE_NAME) --region $(REGION) --platform managed

run-revisions:
	gcloud run revisions list --service $(CLOUD_RUN_SERVICE_NAME) --region $(REGION) --platform managed
