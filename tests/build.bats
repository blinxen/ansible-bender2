#!/usr/bin/env bats
#
# tests for ansible-bender2 build
#

load helpers

##############################################################################
# Bats specific methods                                                      #
##############################################################################

setup() {
    IMAGE_NAME="ab2-test-${BATS_TEST_NUMBER}"
    remove_image "${IMAGE_NAME}"
}

teardown() {
    remove_image "${IMAGE_NAME}"
    remove_image "${IMAGE_NAME}-failed.*"
    buildah rm --all &>/dev/null || true
    buildah rmi --prune &>/dev/null || true
}

##############################################################################
# Local helpers                                                              #
##############################################################################

run_build() {
    local expected_code="${1:-0}"
    shift

    run "${BINARY}" build "$@" "${PLAYBOOK_FILE}"
    if [[ "$status" -ne "$expected_code" ]]; then
        echo "Output:"
        echo "$output"
    fi

    [ "$status" -eq "$expected_code" ]
}

run_build_successfully() {
    run_build 0 "$@"
}

run_build_failing() {
    run_build 1 "$@"
}

write_playbook() {
    local content="$1"
    PLAYBOOK_FILE="${BATS_TEST_TMPDIR}/playbook.yaml"
    echo "$content" > "$PLAYBOOK_FILE"
}

##############################################################################
# Start tests                                                                #
##############################################################################

@test "basic build succeeds and produces an image" {
    # Create a temp file on the host that we'll mount into the build container
    local host_file="${BATS_TEST_TMPDIR}/bats-mount-test.txt"
    echo "hello-from-host\n" > "${host_file}"
    write_playbook "
- hosts: all
  vars:
    ansible_bender:
      base_image: 'python:3.14'
      working_container:
        volumes:
          - ${host_file}:/mnt/test.txt
      target_image:
        name: ${IMAGE_NAME}
        entrypoint:
          - /bin/sh
          - -c
        cmd:
          - --serve
        user: nobody
        environment:
          MY_APP: bats-test
          DEBUG: 'false'
        labels:
          app: ansible-bender2
          suite: bats
        working_dir: /app
        annotations:
          org.opencontainers.image.source: https://example.com
        ports:
          - '8080'
          - '443'
  tasks:
    - name: Echo
      command: echo hello
    - name: Read mounted file
      command: cat /mnt/test.txt
      register: content
    - name: Assert content
      assert:
        that:
          - \"'hello-from-host' in content.stdout\"
"
    run_build_successfully
    image_exists "${IMAGE_NAME}"
    image_not_exists "${IMAGE_NAME}-failed.*"
    image_has_env "${IMAGE_NAME}" "MY_APP=bats-test"
    image_has_env "${IMAGE_NAME}" "DEBUG=false"
    image_has_label "${IMAGE_NAME}" "app" "ansible-bender2"
    image_has_label "${IMAGE_NAME}" "suite" "bats"
    image_has_annotation "${IMAGE_NAME}" "org.opencontainers.image.source" "https://example.com"
    image_has_user "${IMAGE_NAME}" "nobody"
    image_has_working_dir "${IMAGE_NAME}" "/app"
    image_has_entrypoint "${IMAGE_NAME}" "/bin/sh"
    image_has_cmd "${IMAGE_NAME}" "--serve"
    image_has_port "${IMAGE_NAME}" '8080'
    image_has_port "${IMAGE_NAME}" '443'
    image_layer_count_equals "${IMAGE_NAME}" 1

    # building the same playbook twice is idempotent
    run_build_successfully
    image_exists "${IMAGE_NAME}"
    image_not_exists "${IMAGE_NAME}-failed.*"
    image_has_env "${IMAGE_NAME}" "MY_APP=bats-test"
    image_has_env "${IMAGE_NAME}" "DEBUG=false"
    image_has_label "${IMAGE_NAME}" "app" "ansible-bender2"
    image_has_label "${IMAGE_NAME}" "suite" "bats"
    image_has_annotation "${IMAGE_NAME}" "org.opencontainers.image.source" "https://example.com"
    image_has_user "${IMAGE_NAME}" "nobody"
    image_has_working_dir "${IMAGE_NAME}" "/app"
    image_has_entrypoint "${IMAGE_NAME}" "/bin/sh"
    image_has_cmd "${IMAGE_NAME}" "--serve"
    image_has_port "${IMAGE_NAME}" "8080"
    image_has_port "${IMAGE_NAME}" "443"
    image_layer_count_equals "${IMAGE_NAME}" 1
}

@test "build with --no-squash produces more than one layer" {
    write_playbook "
- hosts: all
  vars:
    ansible_bender:
      base_image: 'python:3.14'
      target_image:
        name: ${IMAGE_NAME}
  tasks:
    - name: First task
      command: echo first
    - name: Second task
      command: echo second
    - name: Third task
      command: echo third
"
    run_build_successfully --no-squash
    image_exists "${IMAGE_NAME}"
    image_not_exists "${IMAGE_NAME}-failed.*"

    count=$(inspect_image "python:3.14" | jq -e ".OCIv1.rootfs.diff_ids | length + 1")
    image_layer_count_equals "${IMAGE_NAME}" $count
}

@test "build saves a failure image when --save-failed-image is set" {
    write_playbook "
- hosts: all
  vars:
    ansible_bender:
      base_image: 'python:3.14'
      target_image:
        name: ${IMAGE_NAME}
  tasks:
    - name: Succeed first
      command: echo before-failure
    - name: Intentional failure
      command: /bin/false
    - name: Never reached
      command: echo after-failure
"
    run_build_failing --create-image-on-failure
    image_exists "${IMAGE_NAME}-failed-.*"
    image_not_exists "${IMAGE_NAME}$"
}
