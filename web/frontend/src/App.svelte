<script>
	import { onMount, onDestroy } from "svelte";
	import * as d3 from "d3";
	let ws;

	function getConfig() {
		return {
			nodes: [
				{ id: "A", type: "producer" },
				{ id: "B", type: "producer" },
				{ id: "B1", type: "producer" },
				{ id: "B2", type: "producer" },
				{ id: "B3", type: "producer" },
				{ id: "C", fx: 250, fy: 250 }, // Positioning C to the right of A and B
				{ id: "D", fx: 450, fy: 250 }, // Positioning D to the right of C
				{ id: "E", fx: 450, fy: 150 }, // Positioning E on top of D
				{ id: "C1", type: "consumer" },
				{ id: "C2", type: "consumer" },
				{ id: "C3", type: "consumer" },
			],
			links: [
				{ source: "A", target: "C" },
				{ source: "B", target: "C" },
				{ source: "B1", target: "C" },
				{ source: "B2", target: "C" },
				{ source: "B3", target: "C" },
				{ source: "C", target: "D" },
				{ source: "E", target: "D" },
				{ source: "C1", target: "D" },
				{ source: "C2", target: "D" },
				{ source: "C3", target: "D" },
			],
		};
	}

	const config = getConfig();

	let producerNodes = config.nodes.filter((node) => node.type === "producer");
	let gapBetweenProducerNodes = 100; // Adjust this gap as you see fit

	let consumerNodes = config.nodes.filter((node) => node.type === "consumer");
	let gapBetweenConsumerNodes = 100; // Adjust this gap as you see fit

	onMount(() => {
		const width = 1024;
		const height = 768;

		const socket = new WebSocket("ws://localhost:8899/ws");

		socket.addEventListener("open", (event) => {
			console.log("Connected to the WebSocket server");
		});

		socket.addEventListener("message", (event) => {
			try {
				const msg = JSON.parse(event.data);
				moveCircleAlongLink(svg, msg.source, msg.target);
				console.log("messages received:", event.data);
			} catch (error) {
				console.error("Failed to parse message:", event.data, error);
			}
		});

		socket.addEventListener("error", (error) => {
			console.error(`WebSocket Error: ${error}`);
		});

		socket.addEventListener("close", (event) => {
			if (event.wasClean) {
				console.log(
					`Closed cleanly, code=${event.code}, reason=${event.reason}`
				);
			} else {
				console.error("Connection died");
			}
		});

		const svg = d3.select("#chart").attr("width", width).attr("height", height);

		const simulation = d3.forceSimulation(config.nodes).force(
			"link",
			d3
				.forceLink(config.links)
				.id((d) => d.id)
				.distance(250)
		);

		const link = svg
			.append("g")
			.selectAll("line")
			.data(config.links)
			.join("line")
			.attr("stroke", "black")
			.attr("stroke-width", 2)
			.attr("stroke-dasharray", "2,2")
			.attr("marker-end", "url(#triangle)");

		const node = svg
			.append("g")
			.selectAll("rect")
			.data(config.nodes)
			.join("rect")
			.attr("width", 100)
			.attr("height", 60)
			.attr("rx", 10)
			.attr("ry", 10)
			.attr("fill", "#dedede")
			.attr("filter", "url(#drop-shadow)")
			.call(
				d3
					.drag()
					.on("start", dragStarted)
					.on("drag", dragged)
					.on("end", dragEnded)
			);

		const filter = svg
			.append("defs")
			.append("filter")
			.attr("id", "drop-shadow")
			.attr("height", "130%");

		filter
			.append("feGaussianBlur")
			.attr("in", "SourceAlpha")
			.attr("stdDeviation", 3)
			.attr("result", "blur");

		filter
			.append("feOffset")
			.attr("in", "blur")
			.attr("dx", 2)
			.attr("dy", 2)
			.attr("result", "offsetBlur");

		const feMerge = filter.append("feMerge");
		feMerge.append("feMergeNode").attr("in", "offsetBlur");
		feMerge.append("feMergeNode").attr("in", "SourceGraphic");

		producerNodes.forEach((node, index) => {
			node.fx = 50; // set them to the left of C node
			node.fy =
				250 +
				(index - Math.floor(producerNodes.length / 2)) *
					gapBetweenProducerNodes;
		});

		consumerNodes.forEach((node, index) => {
			node.fx = 600; // set them to the right of D node
			node.fy =
				250 +
				(index - Math.floor(consumerNodes.length / 2)) *
					gapBetweenConsumerNodes;
		});

		simulation.on("tick", () => {
			link
				.attr("x1", (d) => d.source.x + 25)
				.attr("y1", (d) => d.source.y + 15)
				.attr("x2", (d) => d.target.x + 25)
				.attr("y2", (d) => d.target.y + 15);

			node.attr("x", (d) => d.x).attr("y", (d) => d.y);
		});

		moveCircleAlongLink(svg, "A", "C", 1000, () => {
			moveCircleAlongLink(svg, "C", "D", 1000, () => {
				moveCircleAlongLink(svg, "D", "C1");
			});
		});

		setTimeout(() => {
			config.nodes.forEach((node) => {
				if (node.type !== "producer") {
					node.fx = node.x;
					node.fy = node.y;
				}
			});
			simulation.alpha(0).restart();
		}, 2000);

		function dragStarted(event, d) {
			if (!event.active) simulation.alphaTarget(0.1).restart();
			d.fx = d.x;
			d.fy = d.y;
		}

		function dragged(event, d) {
			d.fx = event.x;
			d.fy = event.y;
		}

		function dragEnded(event, d) {
			simulation.stop();
		}

		// Arrowheads
		svg
			.append("svg:defs")
			.selectAll("marker")
			.data(["triangle"])
			.enter()
			.append("svg:marker")
			.attr("id", String)
			.attr("viewBox", "0 -5 10 10")
			.attr("refX", 10)
			.attr("refY", -1.5)
			.attr("markerWidth", 5)
			.attr("markerHeight", 5)
			.attr("orient", "auto")
			.append("svg:path")
			.attr("d", "M0,-5L10,0L0,5");

		function moveCircleAlongLink(
			svg,
			sourceNodeId,
			targetNodeId,
			duration = 1000,
			callback
		) {
			const movingCircle = svg
				.append("circle")
				.attr("r", 5) // Radius of the circle
				.attr("fill", "red"); // Color of the circle

			const sourceNode = config.nodes.find((node) => node.id === sourceNodeId);
			const targetNode = config.nodes.find((node) => node.id === targetNodeId);

			const startTime = Date.now();

			function moveCircle() {
				if (!sourceNode || !targetNode) return;

				const elapsedTime = Date.now() - startTime;
				const t = elapsedTime / duration;

				if (t > 1) {
					movingCircle.remove();
					if (callback) callback();
				} else {
					const dx = targetNode.x - sourceNode.x;
					const dy = targetNode.y - sourceNode.y;

					movingCircle
						.attr("cx", sourceNode.x + t * dx + 25)
						.attr("cy", sourceNode.y + t * dy + 15);

					requestAnimationFrame(moveCircle);
				}
			}

			moveCircle();
		}
	});

	onDestroy(() => {
		clearInterval(ws);
	});
</script>

<svg id="chart" />

<style>
	/* Add styles if needed */
</style>
