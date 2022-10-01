test:
	echo test

zinc:
	mkdir -p data
	docker run -d -v $(CURDIR):/data -e ZINC_DATA_PATH="/data" -p 4080:4080 \
		-e ZINC_FIRST_ADMIN_USER=admin \
		-e ZINC_FIRST_ADMIN_PASSWORD=Complexpass#123 \
		--name zinc \
		--network zinc_default \
		public.ecr.aws/zinclabs/zinc:latest

index:
	docker run -t -v $(CURDIR):/data -w /data \
		--network zinc_default alpine sh -c '\
		apk add curl && \
		curl -X POST http://zinc:4080/api/index -u admin:Complexpass#123 -H "Content-Type: application/json" -d @create_index.json'

upd:
	docker run -t -v $(CURDIR):/data -w /data \
		--network zinc_default alpine sh -c '\
		apk add curl && \
		curl -X PUT http://zinc:4080/api/thing/_mapping -u admin:Complexpass#123 -H "Content-Type: application/json" -d @update_index.json'


doc:
	docker run -t -v $(CURDIR):/data -w /data \
		--network zinc_default alpine sh -c '\
		apk add curl && \
		curl -X POST http://zinc:4080/api/thing/_doc -u admin:Complexpass#123 -H "Content-Type: application/json" -d @create_doc.json'
