package main

import (
        "context"
        "fmt"
        corev1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
        "k8s.io/client-go/kubernetes/scheme"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/tools/clientcmd"
)

func main() {
        fmt.Println("Prepare config object.")


        config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
        if err != nil {
                panic(err)
        }

        config.APIPath = "api"
        config.GroupVersion = &corev1.SchemeGroupVersion
        config.NegotiatedSerializer = scheme.Codecs

        fmt.Println("Init RESTClient.")

        restClient, err := rest.RESTClientFor(config)
        if err != nil {
                panic(err)
        }

        result := &corev1.PodList{}
        if err := restClient.
                Get().
                Namespace("kube-system").
                Resource("pods").
                VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
                Do(context.TODO()).
                Into(result); err != nil {
                panic(err)
        }

        fmt.Println("Print kube-system listed pods.")


        for _, d := range result.Items {
              // fmt.Println(d.Name, d.Status.Phase)
        t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name","Status"})
	t.AppendRows([]table.Row{
		{d.Name, d.Status.Phase},
	})
	t.AppendSeparator()
	t.Render()
        }
}
