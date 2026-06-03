package bpmn

import (
	"context"
	"testing"
)

func BenchmarkBPMN_Parse(b *testing.B) {
	data := []byte(testBPMN)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ParseBPMN(data)
		if err != nil {
			b.Fatalf("ParseBPMN failed: %v", err)
		}
	}
}

func BenchmarkBPMNEngine_Lifecycle(b *testing.B) {
	pp, err := ParseBPMN([]byte(testBPMN))
	if err != nil {
		b.Fatalf("ParseBPMN failed: %v", err)
	}

	engine := NewEngine(pp, nil)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		instance, err := engine.StartInstance("instance-bench", map[string]interface{}{
			"valid": true,
		})
		if err != nil {
			b.Fatalf("StartInstance failed: %v", err)
		}

		// step 1: start -> task1
		if err := engine.Step(ctx, instance); err != nil {
			b.Fatalf("Step failed: %v", err)
		}
		// step 2: task1 -> gateway
		if err := engine.Step(ctx, instance); err != nil {
			b.Fatalf("Step failed: %v", err)
		}
		// step 3: gateway -> end_ok
		if err := engine.Step(ctx, instance); err != nil {
			b.Fatalf("Step failed: %v", err)
		}
		// step 4: end_ok -> completed
		if err := engine.Step(ctx, instance); err != nil {
			b.Fatalf("Step failed: %v", err)
		}
	}
}
