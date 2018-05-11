package kubeclient

import (
	"errors"
	"os"
	"time"

	"git.workshop21.ch/go/abraxas/logging"
	"git.workshop21.ch/workshop21/ba/operator/configuration"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeClient struct {
	KubeConfig         *rest.Config
	SvcConfig          *configuration.Config
	Clientset          *kubernetes.Clientset
	Env                string
	SelectorListOption metav1.ListOptions
}

func GetKubeClient(kc *KubeClient) (*KubeClient, error) {
	if kc == nil {
		cfg, err := configuration.ReadConfig(nil)
		if err != nil {
			return nil, err
		}
		ns := os.Getenv("NAMESPACE")
		return CreateKubeClient(cfg, ns)
	}
	return kc, nil
}

// CreateKubeClient by reading current cluster
func CreateKubeClient(config *configuration.Config, namespace string) (*KubeClient, error) {
	kubeclient, err := CreateKubeClientWithoutSvcConfig(namespace)
	if err != nil {
		return nil, err
	}
	kubeclient.SvcConfig = config
	return kubeclient, nil
}

func (kc *KubeClient) KillOnePodOf(selector string) error {
	kc.SelectorListOption = metav1.ListOptions{LabelSelector: selector}
	pods, err := kc.Clientset.CoreV1().Pods(kc.SvcConfig.RookNamespace).List(kc.SelectorListOption)
	if err != nil {
		return err
	}
	oldestPod, err := getOldestPod(pods.Items)
	if err != nil {
		logging.WithError("BA-OPERATOR-PODKILLER-002", err).Error("Pod could not be found.")
		return err
	}
	logging.WithID("BA-OPERATOR-PODKILLER-001").Info("Name: ", oldestPod.Name)
	go killPod(oldestPod, kc)
	return nil
}

func killPod(pod *v1.Pod, kc *KubeClient) {
	time.Sleep(10 * time.Second)
	err := kc.Clientset.CoreV1().Pods(kc.SvcConfig.RookNamespace).Delete(pod.Name, &metav1.DeleteOptions{})
	if err != nil {
		logging.WithError("BA-OPERATOR-PODKILLER-002", err).Error("Pod could not be killed.")

	}

}

func getOldestPod(pods []v1.Pod) (*v1.Pod, error) {
	if len(pods) == 0 {
		return nil, errors.New("Empty PodList provided")
	}
	oldestPod := v1.Pod{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Time{Time: time.Now()}}}
	for _, pod := range pods {
		if oldestPod.CreationTimestamp.Unix() > pod.CreationTimestamp.Unix() {
			oldestPod = pod
		}
	}

	return &oldestPod, nil
}

// CreateKubeClientWithoutSvcConfig
func CreateKubeClientWithoutSvcConfig(namespace string) (*KubeClient, error) {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := GetClientSet(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &KubeClient{
		KubeConfig: kubeConfig,
		Clientset:  clientset,
		Env:        namespace,
	}, nil
}

// GetClientSet made public
func GetClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}

// GetPodsOfStatefulSet(selector) returns a map of aerodb replicasets
// func (kc *KubeClient) GetPodsOfStatefulSet(selector string) ([]*model.Pod, error) {
// 	statefulSetOptionList := metav1.ListOptions{LabelSelector: selector}
// 	stetefulSets, err := kc.Clientset.AppsV1beta1().StatefulSets(kc.Env).List(statefulSetOptionList)
// 	if err != nil {
// 		logging.WithError("AS-OPERATOR-iuho876b5", err).Error("error retrieving statefulsets from clientset")
// 		return nil, err
// 	}
// 	time.Sleep(10 * time.Second)

// 	var AeroRC []*model.Pod
// 	for _, stetefulSet := range stetefulSets.Items {
// 		statefulSetName := stetefulSet.GetName()
// 		pods, err := kc.getPods(statefulSetName)
// 		if err != nil {
// 			logging.WithError("AS-OPERATOR-hiu2f", err).Error("error retrieving pods from clientset")
// 			return nil, err
// 		}
// 		if len(pods.Items) > 0 {
// 			pod := &model.Pod{PodName: pods.Items[0].Name, PodIP: pods.Items[0].Status.PodIP, StatefulSet: statefulSetName}
// 			AeroRC = append(AeroRC, pod)
// 		}
// 	}
// 	return AeroRC, err
// }

// func (kc *KubeClient) getPods(statefulSetName string) (*v1.PodList, error) {
// 	dbSelectorOptionList := metav1.ListOptions{LabelSelector: fmt.Sprintf("%v ,app=%v", kc.SvcConfig.DBSelector, statefulSetName)}
// 	return kc.Clientset.Core().Pods(kc.Env).List(dbSelectorOptionList)
// }

// func (kc *KubeClient) GetIPOfPod(name string) (string, error) {
// 	pods, err := kc.Clientset.Core().Pods(kc.Env).List(kc.DBSelectorListOption)
// 	if err != nil {
// 		return "", err
// 	}
// 	for _, pod := range pods.Items {
// 		if pod.Name == name {
// 			return pod.Status.PodIP, nil
// 		}
// 	}
// 	return "", nil
// }

// func getClientset(config *rest.Config) (*kubernetes.Clientset, error) {
// 	// creates the clientset
// 	return kubernetes.NewForConfig(config)
// }

// func (kc *KubeClient) GetAeroClusters(config *configuration.Config) (map[string]model.AerospikeCluster, error) {
// 	logging.WithID("AS-OPERATOR-liq34hro8347").Debug("Started to read kubernetes cluster")
// 	aeroClusters := make(map[string]model.AerospikeCluster)
// 	statefulSets, err := kc.Clientset.AppsV1beta1().StatefulSets(kc.Env).List(kc.DBSelectorListOption)
// 	if err != nil {
// 		logging.WithError("AS-OPERATOR-8g234oh87q34fboz", err).Error("error retrieving statefulsets from clientset")
// 		return nil, err
// 	}
// 	logging.WithID("AS-OPERATOR-0z798230h8f4b8o").Debug("start to loop over statefulsets")
// 	for _, statefulSet := range statefulSets.Items {
// 		pods, err := kc.getPods(statefulSet.Name)
// 		if err != nil {
// 			return nil, err
// 		}
// 		aeroCluster := model.AerospikeCluster{Namespace: statefulSet.Namespace}
// 		for _, pod := range pods.Items {
// 			if len(pod.Status.PodIP) < 7 {
// 				continue
// 			}
// 			aeroCluster.PodsInCluster = append(aeroCluster.PodsInCluster, pod.Status.PodIP)
// 		}
// 		aeroClusters[statefulSet.Name] = aeroCluster
// 	}
// 	logging.WithID("AS-OPERATOR-h34h78f4ob8234ho").Debug("finish to loop over statefulsets")
// 	return aeroClusters, err
// }

// GetPods returns an array of strings containing the Subgroups
// func (kc *KubeClient) GetPods(config *configuration.Config) ([]string, error) {
// 	var pods []string

// 	podlist, err := kc.GetPodsOfStatefulSet(config.Selector)
// 	for _, pod := range podlist {
// 		pods = append(pods, pod.StatefulSet)
// 	}
// 	return pods, err
// }

// GetStatefulsetsAsString returns a String of all AeroRS prefixes seperated by comma
// func (kc *KubeClient) GetStatefulsetsAsString(config *configuration.Config) (string, error) {
// 	pods, err := kc.GetPods(config)
// 	return strings.Join(pods, ","), err
// }
