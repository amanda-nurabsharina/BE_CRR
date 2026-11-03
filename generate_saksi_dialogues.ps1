$lines = Get-Content 'c:\Jody\Product\Court Recording Report\BE_CRR\real_bas_saksi_text.txt'
$currentPart = 1
$parts = @{ 1 = @(); 2 = @() }

foreach ($l in $lines) {
    $l = $l.Trim()
    if ($l -match '^PART (\d)') {
        $currentPart = [int]$matches[1]
        continue
    }
    if ($l -match '^([^:]+):\s*(.*)$') {
        $spk = $matches[1].Trim()
        $txt = $matches[2].Trim()
        $parts[$currentPart] += [PSCustomObject]@{ Speaker = $spk; Text = $txt }
    }
}

# Part 1: ~00:00 to 07:30 (450,000ms), Part 2: ~07:30 to 18:37 (1,117,000ms)
$partDurations = @{
    1 = @{ Start = 0; End = 450000 };
    2 = @{ Start = 450000; End = 1117000 }
}

$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine('package usecase')
[void]$sb.AppendLine('')
[void]$sb.AppendLine('var RealCourtSaksiSegments = []RealDialogueSegment{')

foreach ($pNum in 1..2) {
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
        if ($spkTag -eq 'Hakim') { $spkTag = 'Hakim Ketua' }
        elseif ($spkTag -eq 'Penuntut Umum') { $spkTag = 'Jaksa Penuntut' }
        elseif ($spkTag -match 'Saksi') { $spkTag = 'Saksi Fakta' }
        elseif ($spkTag -match 'Penasihat Hukum') { $spkTag = 'Penasihat Hukum' }

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

Set-Content -Path 'c:\Jody\Product\Court Recording Report\BE_CRR\internal\usecase\real_court_saksi_dialogues.go' -Value $sb.ToString() -Encoding utf8
Write-Output "Successfully generated real_court_saksi_dialogues.go"

