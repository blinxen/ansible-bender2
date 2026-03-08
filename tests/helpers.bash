BINARY=${BINARY:-ansible-bender2}

remove_image() {
    local image="$1"
    local match="$(buildah images --format '{{.Name}}:{{.Tag}}' 2>/dev/null | grep -E "${image}" || true)"
    buildah rmi "$match" &>/dev/null || true
}

image_exists() {
    local image="$1"
    local match="$(buildah images --format '{{.Name}}:{{.Tag}}' 2>/dev/null | grep -E "${image}" || true)"
    [[ -n "$match" ]]
}

image_not_exists() {
    local image="$1"
    ! image_exists "$image"
}

inspect_image() {
    local image="$1"
    buildah inspect --type=image "$image"
}

image_has_label() {
    local image="$1"
    local key="$2"
    local value="$3"
    inspect_image "$image" | jq -e "(.OCIv1.config.Labels // {}) | .[\"$key\"] == \"$value\""
}

image_has_annotation() {
    local image="$1"
    local key="$2"
    local value="$3"
    inspect_image "$image" | jq -e ".Manifest | fromjson | (.annotations // {}) | .[\"$key\"] == \"$value\""
}

image_has_env() {
    local image="$1"
    local pair="$2"
    inspect_image "$image" | jq -e "(.OCIv1.config.Env // []) | any(. == \"$pair\")"
}

image_has_user() {
    local image="$1"
    local value="$2"
    inspect_image "$image" | jq -e ".OCIv1.config.User == \"$value\""
}

image_has_working_dir() {
    local image="$1"
    local value="$2"
    inspect_image "$image" | jq -e ".OCIv1.config.WorkingDir == \"$value\""
}

image_has_entrypoint() {
    local image="$1"
    local value="$2"
    inspect_image "$image" | jq -e "(.OCIv1.config.Entrypoint // []) | any(. == \"$value\")"
}

image_has_cmd() {
    local image="$1"
    local value="$2"
    inspect_image "$image" | jq -e "(.OCIv1.config.Cmd // []) | any(. == \"$value\")"
}

image_has_port() {
    local image="$1"
    local value="$2"
    inspect_image "$image" | jq -e "(.OCIv1.config.ExposedPorts // {}) | has(\"$value\")"
}

image_layer_count_equals() {
    local image="$1"
    local count="$2"
    inspect_image "$image" | jq -e ".OCIv1.rootfs.diff_ids | length == $count"
}
