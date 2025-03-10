package cmd

import (
	"context"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// rootCmd represents the base command when called without any subcommands
var (
	cf            *genericclioptions.ConfigFlags
	validPathType = regexp.MustCompile(`(?i)^/[[:alnum:]\_\-/]*$`)
	rootCmd       = &cobra.Command{
		Use:   "ingress-validator",
		Short: "Validate ingress Path",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := newClient()
			if err != nil {
				return err
			}
			ctx := context.TODO()
			ingresses, err := cl.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
			if err != nil {
				return err
			}

			for _, ing := range ingresses.Items {
				for _, rule := range ing.Spec.Rules {
					if rule.HTTP != nil {
						for _, path := range rule.HTTP.Paths {
							if path.Path == "" {
								continue
							}
							if path.PathType == nil || *path.PathType != networkingv1.PathTypeImplementationSpecific {
								if isValid := validPathType.MatchString(path.Path); !isValid {
									cmd.Printf(
										"Ingress: %s/%s with host: %s is not valid",
										ing.GetNamespace(),
										ing.GetName(),
										rule.Host,
									)
								}
							}
						}
					}
				}
			}
			return nil
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cf = genericclioptions.NewConfigFlags(true)
	cf.AddFlags(rootCmd.Flags())
}

func newClient() (*kubernetes.Clientset, error) {
	// try to find a cube config
	cfg, err := cmdutil.NewFactory(cf).ToRESTConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(cfg)
}
