# tailscale-route-tiler

This helper tool generates a list of subnets for TailScale and runs the TailScale command to update the routes.

This tool uses a YAML configuration file; you can find an example in the repo. The Client ID and Tailscale key are required.

## Usage

```bash
  tailscale-route-tiler run -c config.yaml
```

## Help

```bash
tailscale-route-tiler -h
tailscale-route-tiler run -h
tailscale-route-tiler get-client-routes -h
```
