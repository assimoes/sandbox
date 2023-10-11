# Image names
PRODUCER_IMG := producer:latest
FORWARDER_IMG := forwarder:latest
MONITOR_IMG := monitor:latest
CONSUMER_IMG := consumer:latest

PRODUCER_A_CTN := producer_a
FORWARDER_CTN := forwarder
CONSUMER_CTN := consumer_a

.PHONY: producer forwarder monitor consumer build

network:
	@echo "Creating network"
	@docker network create sandbox

b_producer:
	@echo "Building producer"
	@docker build -t $(PRODUCER_IMG) -f producer/Dockerfile .

r_producer:
	@echo "Running producer"
	@docker run -d --name producer_a --network=sandbox -e FORWARDER_URL='http://forwarder:8889' -e EXTERNAL_NAME='producer_a' -e EXTERNAL_PORT=8888 $(PRODUCER_IMG)

b_forwarder:
	@echo "Building forwarder"
	@docker build -t $(FORWARDER_IMG) -f forwarder/Dockerfile .

r_forwarder:
	@echo "Running forwarder"
	@docker run -d --name forwarder --network=sandbox -e KAFKA_BROKER='broker:9092' $(FORWARDER_IMG)


b_monitor:
	@echo "Building monitor"
	@docker build -t $(MONITOR_IMG) -f monitor/Dockerfile .

r_monitor:
	@echo "Running monitor"
	@docker run -d --name monitor --network=sandbox -e KAFKA_BROKER='broker:9092' $(CONSUMER_IMG)

b_consumer:
	@echo "Building consumer"
	@docker build -t $(MONITOR_IMG) -f consumer/Dockerfile .

r_consumer:
	@echo "Running consumer"
	@docker run -d --name consumer_a --network=sandbox -e KAFKA_BROKER='broker:9092' $(CONSUMER_IMG)


build: b_producer b_forwarder b_monitor
	@echo "Docker images built successfully"

run: network r_producer r_forwarder r_monitor
	@echo "Docker containers running successfully"

clean:
	@docker rmi -f $(PRODUCER_IMG) $(FORWARDER_IMG) $(MONITOR_IMG) $(CONSUMER_IMG)
	@docker rm -f $(PRODUCER_A_CTN) $(FORWARDER_CTN) $(MONITOR_CTN) $(CONSUMER_CTN)
	@docker network rm sandbox

