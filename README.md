# XO-Packs-Backend

# Dev build
1. export ENV=dev
2. docker-compose -f docker-compose.$ENV.yml up --build

# Prod build
1. export ENV=prod
2. docker-compose -f docker-compose.$ENV.yml up --build

# Generate swagger docs
1. swag init --parseDependency --parseInternal
