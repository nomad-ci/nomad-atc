# ATC for Nomad

This is a version of [atc](https://github.com/concourse/atc) that retrofits the
worker interface to run containers within a [nomad](https://nomadproject.io)
cluster rather than requiring the management of an explicit worker pool via
[tsa](https://github.com/concourse/tsa).

## Architecture

The primary injection point is
[worker.WorkerProvider](https://godoc.org/github.com/concourse/atc/worker#WorkerProvider).
Our implementation of this interface returns variable implementations of the
Worker, Container, Process, and Volume interfaces to provide versions that can run and
shuffle data to jobs running in the nomad cluster.

### Driver

In order to have the containers perform various tasks related to the concourse
workflows, we inject a special linux driver binary into each task and have nomad
execute that rather than the program concourse wants directly.

This driver takes care of the following duties:

* Capturing the output and error streams from the program and sending them to
  concourse.
* Sending data meant for stdin from concourse to the program.
* Sending signal requests from concourse to the program.
* Downloading volumes meant as inputs for tasks as tars and putting them into
  the filesystem of the container.
* For get and tasks with outputs, waiting to requests of files from their
  filesystems and sending them up to concourse.

### Execution

Execution is fairly straightforward. Nomad provides the ability to start jobs as
docker containers and most (all?) concourse resources and task are specified as
docker images. The driver is run in the context of docker container booted from
the requested image. 

### Volumes

Volumes are the trickiest part because concourse assumes the usage of
baggageclaim and the ability for a volume to outlive a container, something that
doesn't exist in nomad.

Input volumes are easier because the driver can simply request the volume data
before it starts the requested program.

Output volumes require the driver to hold the container alive after the program
has run and wait for any requests for the files it has available in the
fiilesystem. When it gets a request, it tars up the data and sends it back to
concourse. The data is then available to be used directly in course as is the
case of reading a task's file, or by another container.

This means that volumes are less efficent that atc mainline atm because the
data has to be shipped around, even if 2 containers are running on the same
instance. 

Future work could allow nomad to utilize docker mounts from the instance to
allow containers the ability to share a directory containing the volume data.
The difficulty with that is coordinating to have the volume data removed from
the instance once the build is done. For now, the shipping of data around isn't
a huge deal, just makes stuff slower than it could be.

One incompatible change that was made is that anytime a directory is being sent,
the tar is sent through gzip as well. This was done because we ship around much
more data than normal atc and wanted to get some savings back.
