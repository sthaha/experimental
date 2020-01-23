package validate

import (
	"context"
	"io/ioutil"

	"github.com/ghodss/yaml"
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("validate")

type Fn func(path string) error

func Pipeline(path string) error {

	log := log.WithName("pipeline").WithValues("file", path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err, "opening file failed")
		return err
	}

	var pipeline tekton.Pipeline
	if err := yaml.Unmarshal(b, &pipeline); err != nil {
		return err
	}

	ctx := context.Background()
	pipeline.SetDefaults(ctx)

	return pipeline.Validate(ctx)
}

func Task(path string) error {
	log := log.WithName("task").WithValues("file", path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err, "opening file failed")
		return err
	}

	var task tekton.Task
	if err = yaml.Unmarshal(b, &task); err != nil {
		log.Error(err, "yaml unmarshal failed")
		return err
	}

	ctx := context.Background()
	task.SetDefaults(ctx)

	// TODO(sthaha, vdemeester): why does validate return Empty FieldError
	err = sanitizeError(task.Validate(ctx))
	if err != nil {
		log.Error(err, "task validation failed")
	}
	return err
}

func ClusterTask(path string) error {
	log := log.WithName("task").WithValues("file", path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err, "opening file failed")
		return err
	}

	var task tekton.ClusterTask
	if err = yaml.Unmarshal(b, &task); err != nil {
		log.Error(err, "yaml unmarshal failed")
		return err
	}

	ctx := context.Background()
	task.SetDefaults(ctx)

	err = sanitizeError(task.Validate(ctx))
	if err != nil {
		log.Error(err, "task validation failed")
	}
	return err
}

func sanitizeError(fe *apis.FieldError) error {
	if fe.Error() != "" {
		return fe
	}
	return nil
}
