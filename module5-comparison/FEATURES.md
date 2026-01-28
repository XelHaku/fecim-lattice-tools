# Module 5: Comparison - Features

## Features

- **Architecture Comparison** — CPU+DRAM vs GPU+HBM vs FeCIM CIM
- **5 Neural Network Workloads** — MNIST to LLM-70B
- **Data Center Scaling** — Chip count, power, cost, CO2 projections
- **Animated Visualizations** — Energy race, market charts, ROI calculator
- **Market Analysis** — $721B addressable market by 2030

## Comparison Metrics

| Metric | CPU+DRAM | GPU+HBM | FeCIM (Est.) |
|--------|----------|---------|--------------|
| Energy/MAC | 1000 pJ | 100 pJ | ~1 pJ |
| Process Node | 5nm | 4nm | 45nm (est.) |
| TDP | 125W | 400W | 5W (est.) |
| Peak TOPS | 1.0 | 100 | 50 (est.) |
| TOPS/Watt | 0.008 | 0.25 | 10 (est.) |

## Workloads

| Network | MACs | Use Case |
|---------|------|----------|
| MNIST | 101K | Digit recognition |
| ResNet-50 | 4B | Image classification |
| BERT-Base | 11B | NLP |
| GPT-2 | 35B | Language model |
| LLM-70B | 140T | Large language model |

## Key Parameters

| Parameter | Value | Source |
|-----------|-------|--------|
| CPU Energy | 1000 pJ/MAC | Verified |
| GPU Energy | 100 pJ/MAC | Verified |
| FeCIM Energy | ~1 pJ/MAC | Unverified (TRL 4) |
| Electricity | $0.10/kWh | Data center cost |
| PUE | 1.5× | Power overhead |
| CO2 Factor | 0.4 kg/kWh | Emissions |

## Market Data (2030 Projections)

| Segment | Value |
|---------|-------|
| NAND Flash | $98B |
| DRAM | $220B |
| AI Semiconductors | $403B |
| **Total** | $721B |
