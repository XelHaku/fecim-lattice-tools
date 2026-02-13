# Scientific Claims Matrix

This matrix maps each reported scientific claim to executable verification artifacts.

| Claim | Test/Script | Evidence Path | Tolerance/Criterion |
|-------|-------------|---------------|---------------------|
| Pr=19.17 µC/cm² (HZO, Materlik 2015) | TestPhysicsRegression_Preisach | validation/testdata/ | ±5% |
| Ec=1.16 MV/cm | TestPhysicsRegression_LK | validation/testdata/ | ±5% |
| MNIST 80% accuracy | TestFullStackMNIST | validation/ | ≥80% |
| Energy 44.94 fJ/cell | TestTransientPulse | shared/physics/ | 10-100 fJ range |
| KCL residual < 1e-12 | TestKirchhoff_KCL | module4-circuits/pkg/arraysim/ | < 1e-12 |
| Preisach discontinuities physical | TestHeadlessISPPContinuityValidation | cmd/fecim-lattice-tools/ | 0 spurious |
| ISPP converges all targets | TestISPPConverges_Preisach | module1-hysteresis/pkg/controller/ | all levels hit |
| Array ISPP with disturb tracking | TestArrayISPP | shared/physics/ | MaxDisturb < 0.3 |

## Known Limits

- VK-1
- VK-2
- VK-3
- L07
- L08
- L09
- L10

