ansible-bender2
===============

This is a Go rewrite of the python CLI tool [`ansible-bender`](https://github.com/ansible-community/ansible-bender/).

Configuration
-------------

The `ansible-bender2` configuration is read directly from the playbook.

Example:

```yaml
- hosts: all
  vars:
    var1: value
    ansible_bender:
      base_image: python:3.14
      working_container:
        volumes:
          - ./README.md:/doc/README.md
        user: root
      target_image:
        name: ansible_bender_example
        labels:
          app: ansible-bender2
          foo: bar
        annotations:
          app: ansible-bender2
          foo: bar
        environment:
          SOME_ENV: value
        entrypoint:
          - /command
        cmd:
          - --flags
        user: not-root
        working_dir: /run
        ports:
          - "8000"
          - "5000"
        volumes:
          - /volume1
          - /volume2
...
```

Usage
-----

### Build

```bash
ansible-bender2 build <PLAYBOOK>
```

The following flags are available:

* `--no-cache`: do not use caching mecahnism
* `--no-fail-image`: do not create the failure image when an error occurs
* `--squash`: squash image to exactly one layer

License
-------

The source code is primarily distributed under the terms of the MIT License.
See LICENSE for details.
