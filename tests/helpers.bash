BINARY=${BINARY:-ansible-bender2}

remove_image() {
    buildah rmi "$1" &>/dev/null || true
}

image_exists() {
    local name="$1"
    local id
    id="$(buildah images -q "$name" 2>/dev/null)"
    [[ -n "$id" ]]
}

inspect_image() {
    local name="$1"
    buildah inspect --type=image "$name"
}

image_has_label_or_annotation() {
    local image="$1" key="$2" value="$3"
    inspect_image "$image" | grep -q "\"${key}\": \"${value}\""
}

image_has_env() {
    local image="$1" pair="$2"
    inspect_image "$image" | grep -q "\"${pair}\""
}

image_has_user() {
    local image="$1" pair="$2"
    inspect_image "$image" | grep -q "\"${pair}\""
}

image_has_working_dir() {
    local image="$1" pair="$2"
    inspect_image "$image" | grep -q "\"${pair}\""
}

image_has_entrypoint() {
    local image="$1" pair="$2"
    inspect_image "$image" | grep -q "\"${pair}\""
}

image_has_cmd() {
    local image="$1" pair="$2"
    inspect_image "$image" | grep -q "\"${pair}\""
}

image_has_port() {
    local image="$1" pair="$2"
    inspect_image "$image" | grep -q "\"${pair}\""
}

image_layer_count() {
    local name="$1"
    buildah inspect --type=image --format '{{len .OCIv1.RootFS.DiffIDs}}' "$name"
}
