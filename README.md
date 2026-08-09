# Zeroed (Zero Daemon)

Zeroed is the control plane for the Zero panel. It employs a "multi-workload"
system where a server can have a different workload type depending on the
purpose. The daemon handles service lifetimes, backups, logging, and all other
aspects about a server. There's two included workloads, one being the deployment
workload utilizing gVisor to isolate arbitrary Dockerfile-based services to
deploy from code instantly and effectively, the other is a game and general-
purpose workload where Pterodactyl-based eggs can be utilized.

Additionally, it hosts a Git smart HTTP server to immediately deploy code and
services from the terminal, an SFTP server to conveniently and quickly share and
transfer files, and while attempting to minimize third-party modules and dependencies.
