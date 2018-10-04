package cloudconfig

const InstanceStorage = `
storage:
  filesystems:
    - name: ephemeral1
      mount:
        device: /dev/nvme0n1
        format: xfs
        wipeFilesystem: true
`
