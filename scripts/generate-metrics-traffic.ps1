<#
.SYNOPSIS
  Dispara muitos pedidos HTTP 404 e 500 contra a API GoReliable para popular Prometheus.

.PARAMETER BaseUrl
  URL base da API (default: http://localhost:8080). No Docker Compose use http://localhost:8080.

.PARAMETER Count404
  Número de pedidos a paths inexistentes (404).

.PARAMETER Count500
  Número de pedidos a GET /debug/simulate-500 (500).

.EXAMPLE
  .\scripts\generate-metrics-traffic.ps1
  .\scripts\generate-metrics-traffic.ps1 -BaseUrl "http://127.0.0.1:8080" -Count404 500 -Count500 100
#>
param(
    [string]$BaseUrl = "http://localhost:8080",
    [int]$Count404 = 300,
    [int]$Count500 = 50
)

$BaseUrl = $BaseUrl.TrimEnd('/')

function Invoke-Quiet {
    param([string]$Uri, [string]$Method = "GET")
    try {
        Invoke-WebRequest -Uri $Uri -Method $Method -UseBasicParsing -TimeoutSec 10 | Out-Null
    } catch {
        # 4xx/5xx geram erro no Invoke-WebRequest — esperado
    }
}

Write-Host "BaseUrl: $BaseUrl"
Write-Host "Gerando $Count404 x 404 (paths inexistentes)..."
for ($i = 1; $i -le $Count404; $i++) {
    Invoke-Quiet -Uri "$BaseUrl/metrics-traffic-404-$i"
}

Write-Host "Gerando $Count500 x 500 (GET /debug/simulate-500)..."
for ($i = 1; $i -le $Count500; $i++) {
    Invoke-Quiet -Uri "$BaseUrl/debug/simulate-500"
}

Write-Host "Extra: 30 x 404 em DELETE /tasks/:id inexistente..."
for ($i = 1; $i -le 30; $i++) {
    Invoke-Quiet -Uri "$BaseUrl/tasks/not-a-real-task-id-$i" -Method DELETE
}

Write-Host "Concluído. Aguarda o scrape do Prometheus (~15s) e consulta as métricas."
