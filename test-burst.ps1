# Define o número de requisições a serem enviadas
$numberOfRequests = 1000

# Define o intervalo de tempo entre cada requisição (em milissegundos)
$intervalBetweenRequests = 200

# Define os cabeçalhos personalizados para as requisições
$headers = @{
    "Accept" = "*/*"
    "User-Agent" = "Thunder Client (https://www.thunderclient.com)"
}

# Define o URL da API alvo
$reqUrl = '127.0.0.1:8080/api/example'

# Loop para enviar uma rajada de requisições
for ($i = 1; $i -le $numberOfRequests; $i++) {
    Write-Host "Enviando requisição número $i"
    
    # Envia a requisição
    $response = Invoke-RestMethod -Uri $reqUrl -Method Get -Headers $headers

    # Exibe o código de status da resposta
    Write-Host "Resposta da requisição $i $($response.StatusCode)"

    # Aguarda o intervalo definido antes de enviar a próxima requisição
    Start-Sleep -Milliseconds $intervalBetweenRequests
}
