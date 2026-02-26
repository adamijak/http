## Golangci-Lint Failures

There are several instances of unchecked error return values in the file `internal/client/client.go` that are causing failures in golangci-lint. Below are the identified lines and suggestions for adding error handling:

1. **SetDeadline Error Handling**:
   - **Current Code**: 
     ```go
     conn.SetDeadline(time.Now().Add(30 * time.Second))
     ```
   - **Suggested Fix**:
     ```go
     err := conn.SetDeadline(time.Now().Add(30 * time.Second))
     if err != nil { 
         return nil, fmt.Errorf("failed to set deadline: %w", err)
     }
     ```

2. **ReadString Error Handling**:
   - **Current Code**: 
     ```go
     line, err := reader.ReadString('\n')
     ```
   - **Suggested Fix**:  
     ```go
     _, err := reader.ReadString('\n')
     if err != nil {
         return "", err
     }
     ```

These changes should be applied to all similar instances throughout the file. Proper error handling will resolve the linter errors and allow CI to pass.