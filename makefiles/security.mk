scan/code: ## scans code for vulnerabilities
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy fs /helm-images

scan/binary: ## scans binary for vulnerabilities
	@docker-compose --project-name trivy -f docker-compose.trivy.yml run --rm trivy fs /helm-images/dist/helm-images_darwin_amd64_v1/helm-images --scanners vuln
