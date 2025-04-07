#!/bin/bash

# Update ArgoCD to use the Helm chart
cat <<EOF | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: whosay
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/tiwariParth/whosay.git
    targetRevision: HEAD
    path: helm/whosay
    helm:
      valueFiles:
        - values.yaml
        - values-dev.yaml  # Can be changed to values-prod.yaml for production
  destination:
    server: https://kubernetes.default.svc
    namespace: whosay
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
EOF

echo "ArgoCD application updated to use Helm chart."
echo "You can check the status with: kubectl get applications -n argocd whosay"
