package controller

import (
	"testing"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
)

func Test_makeAnnotations_NoKeys(t *testing.T) {
	annotationVal := `{"name":"","image":""}`

	spec := v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{},
	}

	annotations := makeAnnotations(&spec)

	if _, ok := annotations["prometheus.io.scrape"]; !ok {
		t.Errorf("wanted annotation " + "prometheus.io.scrape" + " to be added")
		t.Fail()
	}
	if val, _ := annotations["prometheus.io.scrape"]; val != "false" {
		t.Errorf("wanted annotation " + "prometheus.io.scrape" + ` to equal "false"`)
		t.Fail()
	}

	if _, ok := annotations[annotationInferenceSpec]; !ok {
		t.Errorf("wanted annotation " + annotationInferenceSpec)
		t.Fail()
	}

	if val, _ := annotations[annotationInferenceSpec]; val != annotationVal {
		t.Errorf("Annotation " + annotationInferenceSpec + "\nwant: '" + annotationVal + "'\ngot: '" + val + "'")
		t.Fail()
	}
}

func Test_makeAnnotations_WithKeyAndValue(t *testing.T) {
	annotationVal := `{"name":"","image":"","annotations":{"key":"value","key2":"value2"}}`

	spec := v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Annotations: map[string]string{
				"key":  "value",
				"key2": "value2",
			},
		},
	}

	annotations := makeAnnotations(&spec)

	if _, ok := annotations["prometheus.io.scrape"]; !ok {
		t.Errorf("wanted annotation " + "prometheus.io.scrape" + " to be added")
		t.Fail()
	}
	if val := annotations["prometheus.io.scrape"]; val != "false" {
		t.Errorf("wanted annotation " + "prometheus.io.scrape" + ` to equal "false"`)
		t.Fail()
	}

	if _, ok := annotations[annotationInferenceSpec]; !ok {
		t.Errorf("wanted annotation " + annotationInferenceSpec)
		t.Fail()
	}

	if val := annotations[annotationInferenceSpec]; val != annotationVal {
		t.Errorf("Annotation " + annotationInferenceSpec + "\nwant: '" + annotationVal + "'\ngot: '" + val + "'")
		t.Fail()
	}
}

func Test_makeAnnotationsDoesNotModifyOriginalSpec(t *testing.T) {
	specAnnotations := map[string]string{
		"test.foo": "bar",
	}
	function := &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:        "testfunc",
			Annotations: specAnnotations,
		},
	}

	expectedAnnotations := map[string]string{
		"prometheus.io.scrape":  "false",
		"test.foo":              "bar",
		annotationInferenceSpec: `{"name":"testfunc","image":"","annotations":{"test.foo":"bar"}}`,
	}

	makeAnnotations(function)
	annotations := makeAnnotations(function)

	if len(specAnnotations) != 1 {
		t.Errorf("length of original spec annotations has changed, expected 1, got %d", len(specAnnotations))
	}

	if specAnnotations["test.foo"] != "bar" {
		t.Errorf("original spec annotation has changed")
	}

	for name, expectedValue := range expectedAnnotations {
		actualValue := annotations[name]
		if actualValue != expectedValue {
			t.Fatalf("incorrect annotation for '%s': \nwant %q,\ngot %q", name, expectedValue, actualValue)
		}
	}
}
