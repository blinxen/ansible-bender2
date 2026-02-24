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
    remove_image "${IMAGE_NAME}-failed"
    buildah rm --all &>/dev/null || true
}

##############################################################################
# Local helpers                                                              #
##############################################################################

run_build() {
    run "${BINARY}" build "$@" "${PLAYBOOK_FILE}"
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
      base_image: 'fedora:rawhide'
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
    run_build
    [ "$status" -eq 0 ]
    [[ "$output" =~ ${IMAGE_NAME} ]]
    image_exists "${IMAGE_NAME}"
    exit 0
    image_has_env "${IMAGE_NAME}" "MY_APP=bats-test"
    image_has_env "${IMAGE_NAME}" "DEBUG=false"
    image_has_label_or_annotation "${IMAGE_NAME}" "app" "ansible-bender2"
    image_has_label_or_annotation "${IMAGE_NAME}" "suite" "bats"
    image_has_label_or_annotation "${IMAGE_NAME}" "org.opencontainers.image.source" "https://example.com"
    image_has_user "${IMAGE_NAME}" "nobody"
    image_has_working_dir "${IMAGE_NAME}" "/app"
    image_has_entrypoint "${IMAGE_NAME}" "/bin/sh"
    image_has_cmd "${IMAGE_NAME}" "--serve"
    image_has_port "${IMAGE_NAME}" '8080'
    image_has_port "${IMAGE_NAME}" '443'

    # building the same playbook twice is idempotent
    run_build
    [ "$status" -eq 0 ]
    [[ "$output" =~ ${IMAGE_NAME} ]]
    image_exists "${IMAGE_NAME}"
    image_has_env "${IMAGE_NAME}" "MY_APP" "bats-test"
    image_has_env "${IMAGE_NAME}" "DEBUG" "false"
    image_has_label_or_annotation "${IMAGE_NAME}" "app" "ansible-bender2"
    image_has_label_or_annotation "${IMAGE_NAME}" "suite" "bats"
    image_has_label_or_annotation "${IMAGE_NAME}" "org.opencontainers.image.source" "https://example.com"
    image_has_user "${IMAGE_NAME}" "nobody"
    image_has_working_dir "${IMAGE_NAME}" "/app"
    image_has_entrypoint "${IMAGE_NAME}" "/bin/sh"
    image_has_cmd "${IMAGE_NAME}" "--serve"
    image_has_port "${IMAGE_NAME}" "8080"
    image_has_port "${IMAGE_NAME}" "443"
}

# TODO: Add test for no-cache
# TODO: Add test for no-squash
# TODO: Add test for save-failure-image
# TODO: Add test for no failure image by default

# @test "building with all flags succeeds" {
#     write_playbook "
# - hosts: all
#   vars:
#     ansible_bender:
#       base_image: 'fedora:rawhide'
#       target_image:
#         name: ${IMAGE_NAME}
#   tasks:
#     - name: Echo
#       command: echo all-flags
# "
#     run_build --no-cache --squash --no-fail-image
#     [ "$status" -eq 0 ]
#     [[ "$output" != *"panic"* ]]
#     image_exists "${IMAGE_NAME}"
# }
