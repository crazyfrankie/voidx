# Errno Code Gen
## Production

When use this tool generating errno code, you should create two file at least for system metadata, such as: `./example/example_metadata.yaml` and `./example/example_common.yaml`.
It will provide the project errno settings and common error code, and then you should create a file for every biz module like `./example/example_user.yaml`.

## Usage
```bash
go run code_gen.go \
     --biz {bizName} \
     --app-name {appName} \
     --app-code {app-code} \
     --import-path {import-path} \
     --output-dir {output-dir} \
     --script-dir {script-dir}
```

## Example
Before use this command, you should rename or touch a new file without "example" prefix in example directory:
```bash
go run code_gen.go \
     --biz user \
     --app-name myapp \
     --app-code 6 \
     --import-path "github.com/crazyfrankie/frx/errorx/code" \
     --output-dir "./generated/user" \
     --script-dir "./example"
```

## Custom Error Code Length
To use custom error code length, add the following section to your metadata.yaml:
```yaml
error_code:
  # Total length of the error code (default: 9)
  total_length: 9
  # Length of app code (default: 1)
  app_length: 1
  # Length of business code (default: 3)
  biz_length: 3
  # Length of sub code (default: 4)
  sub_length: 4
```
The tool will automatically calculate the error codes based on these lengths.