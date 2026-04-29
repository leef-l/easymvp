with open('apps/desktop/src/renderer/app/shell.tsx', 'r', encoding='utf-8') as f:
    lines = f.readlines()
for i, line in enumerate(lines):
    if 'routes.acceptance' in line:
        lines[i] = '      { to: routes.acceptance, label: t("nav.acceptance"), icon: "\u2705" },\n'
with open('apps/desktop/src/renderer/app/shell.tsx', 'w', encoding='utf-8') as f:
    f.writelines(lines)
print('done')
