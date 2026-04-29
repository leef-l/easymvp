with open(r'C:\Users\Public\project\brain\cmd\brain\cmd_serve.go', 'r', encoding='utf-8') as f:
    content = f.read()

idx = content.find('// POST /v1/runs')
if idx >= 0:
    insert = (
        '\t\t// POST /v1/contracts/execute - direct contract execution on a sidecar brain\n'
        '\t\tmux.HandleFunc("/v1/contracts/execute", func(w http.ResponseWriter, r *http.Request) {\n'
        '\t\t\tif r.Method != http.MethodPost {\n'
        '\t\t\t\thttp.Error(w, "{\\\"error\\\":\\\"method not allowed\\\"}", http.StatusMethodNotAllowed)\n'
        '\t\t\t\treturn\n'
        '\t\t\t}\n'
        '\t\t\thandleContractExecute(w, r, mgr)\n'
        '\t\t})\n'
        '\n'
    )
    content = content[:idx] + insert + content[idx:]
    with open(r'C:\Users\Public\project\brain\cmd\brain\cmd_serve.go', 'w', encoding='utf-8') as f:
        f.write(content)
    print('OK')
else:
    print('NOT FOUND')
