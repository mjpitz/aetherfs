job "aetherfs-us" {
  multiregion {
    strategy {
      max_parallel = 1
      on_failure = "fail_all"
    }

    region "ap" {
      // ...
    }

    region "us" {
      // ...
    }

    region "eu" {
      // ...
    }
  }

  // group "canary" { ... }

  group "main" {
    task "aetherfs" {
      driver = "exec"
      config {
        command = "./local/bin/aetherfs"
        args = []
      }

      artifact {
        source = "https://github.com/mjpitz/aetherfs/releases/download/v21.10.0/aetherfs_linux_amd64.tar.gz"
        destination = "local/bin"
        options {
          checksum = "sha256:ec5d6268222bc8198058b5ded00f72eb00a478a96fa2e74f098668d556be6529"
        }
      }

      service {
        tags = ["http", "nfs"]

        check {
          type = "tcp"
          port = "http"
          interval = "10s"
          timeout = "2s"
        }

        check {
          type = "tcp"
          port = "nfs"
          interval = "10s"
          timeout = "2s"
        }
      }

      env {}

      resources {}
    }

    network {
      port "http" {
        to = 8080
      }

      port "nfs" {
        to = 2049
      }
    }
  }
}