$lines = Get-Content 'c:\Jody\Product\Court Recording Report\BE_CRR\real_bas_text.txt'
$currentPart = 1
$parts = @{ 1 = @(); 2 = @(); 3 = @() }

foreach ($l in $lines) {
    $l = $l.Trim()
    if ($l -match '^BAGIAN (\d)') {
        $currentPart = [int]$matches[1]
        continue
    }
    if ($l -match '^([^:]+):\s*(.*)$') {
        $spk = $matches[1].Trim()
        $txt = $matches[2].Trim()
        $parts[$currentPart] += [PSCustomObject]@{ Speaker = $spk; Text = $txt }
    }
}

$partDurations = @{
    1 = @{ Start = 0; End = 423000 };
    2 = @{ Start = 423000; End = 803000 };
    3 = @{ Start = 803000; End = 1050000 }
}

$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine('package usecase')
[void]$sb.AppendLine('')
[void]$sb.AppendLine('type RealDialogueSegment struct {')
[void]$sb.AppendLine('	Speaker string')
[void]$sb.AppendLine('	Text    string')
[void]$sb.AppendLine('	StartMs int')
[void]$sb.AppendLine('	EndMs   int')
[void]$sb.AppendLine('}')
[void]$sb.AppendLine('')
[void]$sb.AppendLine('var RealCourtVerbatimSegments = []RealDialogueSegment{')

foreach ($pNum in 1..3) {
    $items = $parts[$pNum]
    $pStart = $partDurations[$pNum].Start
    $pEnd = $partDurations[$pNum].End
    $pSpan = $pEnd - $pStart

    $totalWeight = 0
    foreach ($it in $items) {
        $w = [Math]::Max(15, $it.Text.Length)
        $totalWeight += $w
    }

    $curr = $pStart
    for ($i = 0; $i -lt $items.Count; $i++) {
        $it = $items[$i]
        $w = [Math]::Max(15, $it.Text.Length)
        $segSpan = [int]($pSpan * ($w / $totalWeight))
        $sEnd = $curr + $segSpan
        if ($i -eq ($items.Count - 1)) { $sEnd = $pEnd }

        $spkTag = $it.Speaker
        if ($spkTag -eq 'Hakim/Penuntut Umum') { $spkTag = 'Penuntut Umum' }
        elseif ($spkTag -eq 'Hakim') { $spkTag = 'Hakim Ketua' }
        elseif ($spkTag -eq 'Terdakwa (Alim)') { $spkTag = 'Terdakwa' }

        $escapedText = $it.Text.Replace('"', '\"')
        [void]$sb.AppendLine('	{')
        [void]$sb.AppendLine("		Speaker: `"$spkTag`",")
        [void]$sb.AppendLine("		Text: `"$escapedText`",")
        [void]$sb.AppendLine("		StartMs: $curr,")
        [void]$sb.AppendLine("		EndMs: $sEnd,")
        [void]$sb.AppendLine('	},')

        $curr = $sEnd
    }
}

[void]$sb.AppendLine('}')

Set-Content -Path 'c:\Jody\Product\Court Recording Report\BE_CRR\internal\usecase\real_court_dialogues.go' -Value $sb.ToString() -Encoding utf8
Write-Output "Successfully generated real_court_dialogues.go"

