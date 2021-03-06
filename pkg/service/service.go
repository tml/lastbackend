package service

import (
	"errors"
	"fmt"
	"github.com/lastbackend/lastbackend/libs/adapter/k8s/converter"
	e "github.com/lastbackend/lastbackend/libs/errors"
	"github.com/lastbackend/lastbackend/libs/interface/k8s"
	"github.com/lastbackend/lastbackend/pkg/service/resource/deployment"
	"github.com/lastbackend/lastbackend/pkg/util/generator"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
)

type Service struct {
	deployment.Deployment
	config *v1beta1.Deployment
}

func Get(client k8s.IK8S, namespace, name string) (*Service, *e.Err) {

	var er error

	detail, er := deployment.Get(client, namespace, name)
	if er != nil {
		return nil, e.New("service").Unknown(er)
	}

	return &Service{*detail, nil}, nil
}

func List(client k8s.IK8S, namespace string) (map[string]*Service, *e.Err) {

	var (
		er          error
		serviceList = make(map[string]*Service)
	)

	detailList, er := deployment.List(client, namespace)
	if er != nil {
		return nil, e.New("service").Unknown(er)
	}

	for _, val := range detailList {
		serviceList[val.ObjectMeta.Name] = &Service{val, nil}
	}

	return serviceList, nil
}

func Create(config interface{}) (*Service, *e.Err) {

	var s = new(Service)

	switch config.(type) {
	case *v1beta1.Deployment:
		s.config = config.(*v1beta1.Deployment)
	case *v1.ReplicationController:
		s.config = converter.Convert_ReplicationController_to_Deployment(config.(*v1.ReplicationController))
	case *v1.Pod:
		s.config = converter.Convert_Pod_to_Deployment(config.(*v1.Pod))
	default:
		return nil, e.New("service").Unknown(errors.New("unknown type config"))
	}

	s.config.Name = fmt.Sprintf("%s-%s", s.config.Name, generator.GetUUIDV4()[0:12])

	return s, nil
}

func Update(client k8s.IK8S, namespace, name string, config *ServiceConfig) *e.Err {

	var er error

	dp, er := client.Extensions().Deployments(namespace).Get(name)
	if er != nil {
		return e.New("service").Unknown(er)
	}

	config.update(dp)

	er = deployment.Update(client, namespace, dp)
	if er != nil {
		return e.New("service").Unknown(er)
	}

	return nil
}

func (s Service) Deploy(client k8s.IK8S, namespace string) (*Service, *e.Err) {

	var er error

	_, er = client.Extensions().Deployments(namespace).Create(s.config)
	if er != nil {
		return nil, e.New("service").Unknown(er)
	}

	detail, er := deployment.Get(client, namespace, s.ObjectMeta.Name)
	if er != nil {
		return nil, e.New("service").Unknown(er)
	}

	return &Service{*detail, nil}, nil
}
