"""
Extract all physics equations and circuit models from Module 4 source code.
"""
import re
import json
from pathlib import Path
from datetime import datetime

# Physics equations catalog
physics_catalog = {
    "DAC_EQUATIONS": [],
    "ADC_EQUATIONS": [],
    "TIA_EQUATIONS": [],
    "CHARGEPUMP_EQUATIONS": [],
    "KIRCHHOFF_LAWS": [],
    "SIGNAL_FLOW": [],
    "TIMING_MODELS": []
}

# Source files to analyze
source_files = [
    "module4-circuits/pkg/peripherals/dac.go",
    "module4-circuits/pkg/peripherals/adc.go",
    "module4-circuits/pkg/peripherals/tia.go",
    "module4-circuits/pkg/peripherals/chargepump.go",
    "module4-circuits/pkg/peripherals/analysis.go"
]

# Documentation files
doc_files = [
    "docs/peripheral-circuits/circuits.CIM-fundamentals.md",
    "docs/peripheral-circuits/circuits.operations.md",
    "docs/peripheral-circuits/MODULE4-PHYSICS-IMPROVEMENTS.md"
]

def extract_equations_from_code(filepath):
    """Extract physics equations from Go source code."""
    equations = []
    
    with open(filepath, 'r') as f:
        content = f.read()
        lines = content.split('\n')
        
        for i, line in enumerate(lines):
            # Look for mathematical operations in comments
            if '//' in line and any(op in line for op in ['=', '*', '/', '+', '-', '^', '√', '∫']):
                equations.append({
                    'line': i + 1,
                    'equation': line.strip(),
                    'file': filepath
                })
            
            # Look for calculations in code
            if any(keyword in line for keyword in ['math.', 'voltage', 'current', 'power', 'energy']):
                # Extract variable assignments with calculations
                if '=' in line and not '==' in line and not '!=' in line:
                    equations.append({
                        'line': i + 1,
                        'equation': line.strip(),
                        'file': filepath
                    })
    
    return equations

def extract_constants(filepath):
    """Extract physical constants and default parameters."""
    constants = {}
    
    with open(filepath, 'r') as f:
        content = f.read()
        
        # Find all numeric constants with units in comments
        pattern = r'(\w+)\s*[:=]\s*([-\d.eE+]+)\s*(?:,\s*)?//.*?([A-Za-z]+)'
        matches = re.finditer(pattern, content)
        
        for match in matches:
            name, value, unit = match.groups()
            constants[name] = {
                'value': value,
                'unit': unit,
                'file': filepath
            }
    
    return constants

def categorize_equation(eq_text):
    """Categorize equation by component."""
    eq_lower = eq_text.lower()
    
    if 'dac' in eq_lower or 'digital' in eq_lower or 'analog' in eq_lower:
        return 'DAC_EQUATIONS'
    elif 'adc' in eq_lower or 'convert' in eq_lower and 'voltage' in eq_lower:
        return 'ADC_EQUATIONS'
    elif 'tia' in eq_lower or 'current' in eq_lower and 'gain' in eq_lower:
        return 'TIA_EQUATIONS'
    elif 'pump' in eq_lower or 'boost' in eq_lower or 'dickson' in eq_lower:
        return 'CHARGEPUMP_EQUATIONS'
    elif 'kirchhoff' in eq_lower or 'kcl' in eq_lower or 'sum' in eq_lower:
        return 'KIRCHHOFF_LAWS'
    elif 'timing' in eq_lower or 'settle' in eq_lower or 'delay' in eq_lower:
        return 'TIMING_MODELS'
    else:
        return 'SIGNAL_FLOW'

# Extract from all source files
print("[DATA] Extracting physics equations from source code...")
all_equations = []
all_constants = {}

for source_file in source_files:
    filepath = f"<local-path>"
    if Path(filepath).exists():
        equations = extract_equations_from_code(filepath)
        constants = extract_constants(filepath)
        
        all_equations.extend(equations)
        all_constants.update(constants)
        
        print(f"[FINDING] Extracted {len(equations)} equations from {source_file}")
    else:
        print(f"[LIMITATION] File not found: {source_file}")

# Categorize equations
for eq in all_equations:
    category = categorize_equation(eq['equation'])
    physics_catalog[category].append(eq)

# Generate summary statistics
print("\n[FINDING] Physics Equation Catalog Summary:")
for category, equations in physics_catalog.items():
    print(f"[STAT:{category.lower()}_count] {len(equations)}")

print(f"\n[STAT:total_equations] {len(all_equations)}")
print(f"[STAT:total_constants] {len(all_constants)}")

# Save to JSON for further analysis
output_path = "<local-path>"
with open(output_path, 'w') as f:
    json.dump({
        'catalog': physics_catalog,
        'constants': all_constants,
        'timestamp': datetime.now().isoformat(),
        'total_equations': len(all_equations)
    }, f, indent=2)

print(f"\n[FINDING] Physics catalog saved to {output_path}")
