package cloudconfig

const Small = `{
  "ignition": {
    "version": "2.2.0",
    "config": {
      "replace": {
        "source": "{{ .S3HTTPURL }}"
      }
    }
  }
}
`
