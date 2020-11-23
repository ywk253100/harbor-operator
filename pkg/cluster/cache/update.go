package cache

import (
	"github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"

	"github.com/goharbor/harbor-operator/pkg/lcm"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// RollingUpgrades reconcile will rolling upgrades Redis sentinel cluster if resource upscale.
// It does:
// - check resource
// - update RedisFailovers CR resource
func (rc *RedisController) RollingUpgrades(cluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	crdClient := rc.DClient.WithResource(redisFailoversGVR).WithNamespace(cluster.Namespace)
	if rc.expectCR == nil {
		return cacheUnknownStatus(), nil
	}

	var (
		actualCR, expectCR redisOp.RedisFailover
	)

	if !IsEqual(actualCR.DeepCopy().Spec, expectCR.DeepCopy().Spec) {
		//msg := fmt.Sprintf(UpdateMessageRedisCluster, cluster.Name)
		//rc.Recorder.Event(cluster, corev1.EventTypeNormal, RedisUpScaling, msg)
		rc.Log.Info(
			"Update Redis resource",
			"namespace", cluster.Namespace, "name", cluster.Name,
		)

		expectCR.ObjectMeta.SetResourceVersion(actualCR.ObjectMeta.GetResourceVersion())
		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&expectCR)
		if err != nil {
			return cacheUnknownStatus(), nil
		}

		_, err = crdClient.Update(&unstructured.Unstructured{Object: data}, metav1.UpdateOptions{})
		if err != nil {
			return cacheUnknownStatus(), err
		} else {
			return nil, nil
		}
	}
	return cacheUnknownStatus(), nil
}

func (rc *RedisController) Update(cluster *v1alpha2.HarborCluster) (*lcm.CRStatus, error) {
	crStatus, err := rc.RollingUpgrades(cluster)
	if err != nil {
		return crStatus, err
	}
	return crStatus, nil
}