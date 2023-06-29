import subprocess
import sys
from typing import List


def run(commands: List[str]):
    print(" ".join(commands))
    subprocess.run(commands, stdout=sys.stdout, shell=True, check=True)


def main():
    run(['kubectl apply -f prepare.yaml'])
    run(['go build -o example .'])
    run(['./example '
         '--engine kube '
         '--kube-config=$HOME/.kube/config '
         '--kube-use-port-forward=true '
         '--kube-executor-image docker.io/cox96de/runner-executor:master '
         '--kube-namespace kube-engine-test-executor '
         ' testdata/hello-world.yaml testdata/gobuild.yaml'
         ])


if __name__ == '__main__':
    main()
