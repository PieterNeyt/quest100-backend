package domain

import (
	"fmt"
	"math/rand"
	"time"
)

const FormulaLength = 8

func GenerateDailyFormula(date time.Time) string {
	seed := int64(date.Year())*10000 + int64(date.Month())*100 + int64(date.Day())
	rng := rand.New(rand.NewSource(seed))

	generators := []func(*rand.Rand) (string, bool){
		generateAddition,
		generateSubtraction,
		generateMultiplication,
		generateDivision,
		generateTwoOperators,
	}

	rng.Shuffle(len(generators), func(i, j int) {
		generators[i], generators[j] = generators[j], generators[i]
	})

	for i := 0; i < 100; i++ {
		for _, gen := range generators {
			if formula, ok := gen(rng); ok {
				return formula
			}
		}
	}

	return "3+9-2=10"
}

func isExactLength(formula string) bool {
	return len([]rune(formula)) == FormulaLength
}

func validate(formula string) bool {
	return isExactLength(formula) && ValidateFormula(formula) == nil
}

func generateAddition(rng *rand.Rand) (string, bool) {
	type split struct{ aMax, bMax, rMin, rMax int }
	splits := []split{
		{9, 9, 1000, 9999},
		{9, 99, 100, 999},
		{9, 999, 10, 99},
		{99, 9, 100, 999},
		{99, 99, 10, 99},
		{99, 999, 1, 9},
		{999, 9, 10, 99},
		{999, 99, 1, 9},
		{9999, 9, 1, 9},
	}
	rng.Shuffle(len(splits), func(i, j int) { splits[i], splits[j] = splits[j], splits[i] })

	for _, s := range splits {
		for attempt := 0; attempt < 20; attempt++ {
			a := randInRange(rng, 1, s.aMax)
			b := randInRange(rng, 1, s.bMax)
			r := a + b
			if r < s.rMin || r > s.rMax {
				continue
			}
			f := fmt.Sprintf("%d+%d=%d", a, b, r)
			if validate(f) {
				return f, true
			}
		}
	}
	return "", false
}

func generateSubtraction(rng *rand.Rand) (string, bool) {
	type split struct{ bMax, rMax, aMin, aMax int }
	splits := []split{
		{9, 9999, 1001, 9999},
		{99, 999, 100, 999},
		{9, 999, 100, 999},
		{9, 99, 10, 99},
		{99, 99, 10, 99},
		{99, 9, 10, 99},
		{999, 9, 10, 99},
		{9, 9, 1, 9},
		{99, 99, 10, 99},
	}
	rng.Shuffle(len(splits), func(i, j int) { splits[i], splits[j] = splits[j], splits[i] })

	for _, s := range splits {
		for attempt := 0; attempt < 20; attempt++ {
			b := randInRange(rng, 1, s.bMax)
			r := randInRange(rng, 1, s.rMax)
			a := b + r
			if a < s.aMin || a > s.aMax {
				continue
			}
			f := fmt.Sprintf("%d-%d=%d", a, b, r)
			if validate(f) {
				return f, true
			}
		}
	}
	return "", false
}

func generateMultiplication(rng *rand.Rand) (string, bool) {
	for attempt := 0; attempt < 200; attempt++ {
		a := randInRange(rng, 1, 999)
		b := randInRange(rng, 1, 999)
		r := a * b
		f := fmt.Sprintf("%d*%d=%d", a, b, r)
		if validate(f) {
			return f, true
		}
	}
	return "", false
}

func generateDivision(rng *rand.Rand) (string, bool) {
	for attempt := 0; attempt < 200; attempt++ {
		b := randInRange(rng, 2, 999)
		r := randInRange(rng, 1, 999)
		a := b * r
		f := fmt.Sprintf("%d/%d=%d", a, b, r)
		if validate(f) {
			return f, true
		}
	}
	return "", false
}

func generateTwoOperators(rng *rand.Rand) (string, bool) {
	ops := []string{"+", "-", "*", "/"}

	for attempt := 0; attempt < 300; attempt++ {
		op1 := ops[rng.Intn(len(ops))]
		op2 := ops[rng.Intn(len(ops))]

		a := randInRange(rng, 1, 9)
		b := randInRange(rng, 1, 9)
		c := randInRange(rng, 1, 9)

		expr := fmt.Sprintf("%d%s%d%s%d", a, op1, b, op2, c)
		result, err := evalExpression(expr)
		if err != nil || result <= 0 || result != float64(int(result)) {
			continue
		}
		f := fmt.Sprintf("%s=%d", expr, int(result))
		if validate(f) {
			return f, true
		}
	}
	return "", false
}

func randInRange(rng *rand.Rand, min, max int) int {
	if max <= min {
		return min
	}
	return min + rng.Intn(max-min+1)
}
