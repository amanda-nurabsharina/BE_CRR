package main

import (
	"log"

	"be_crr/internal/domain/entity"
	"be_crr/internal/infrastructure/config"
	"be_crr/internal/infrastructure/database"
	"be_crr/internal/usecase"

	"github.com/google/uuid"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}

	log.Println("1. Updating Court Case to real trial details...")
	var courtCase entity.CourtCase
	if err := db.First(&courtCase).Error; err != nil {
		log.Fatalf("Find case failed: %v", err)
	}

	courtCase.CaseNumber = "51/Pid.B/2026/PN.Tlm"
	courtCase.ClientName = "Alim (Terdakwa)"
	courtCase.CaseCategory = "Pencurian Kendaraan Bermotor (Pasal 362 KUHP)"
	courtCase.CourtName = "Pengadilan Negeri Tilamuta (Kabupaten Boalemo)"
	courtCase.CounterPartyName = "Jaksa Penuntut Umum Kejari Boalemo"
	courtCase.JudgePanelNames = "Dr. Wahyu Iman S.H., M.H. (Ketua), Siti Rahma S.H., Hendra S.H."
	if err := db.Save(&courtCase).Error; err != nil {
		log.Fatalf("Update case failed: %v", err)
	}

	log.Println("2. Updating Court Session...")
	var session entity.CourtSession
	if err := db.Where("court_case_id = ?", courtCase.ID).First(&session).Error; err == nil {
		session.SessionTitle = "Pemeriksaan Keterangan Terdakwa (Pencurian Sepeda Motor Yamaha X-Ride)"
		_ = db.Save(&session)
	}

	log.Println("3. Updating Audio Evidence...")
	var audio entity.AudioEvidence
	if err := db.Where("original_filename LIKE ?", "%Terdakwa%").First(&audio).Error; err == nil {
		audio.DurationSeconds = 1050
		_ = db.Save(&audio)
	} else {
		_ = db.First(&audio)
	}

	log.Println("4. Updating Transcript & 226 Verbatim Segments...")
	var transcript entity.Transcript
	if err := db.Where("audio_evidence_id = ?", audio.ID).First(&transcript).Error; err != nil {
		transcript = entity.Transcript{
			TenantID:        courtCase.TenantID,
			AudioEvidenceID: audio.ID,
			GenerationMode:  "auto_asr",
			Status:          "finalized",
			VersionID:       2,
		}
		_ = db.Create(&transcript)
	} else {
		transcript.GenerationMode = "auto_asr"
		transcript.Status = "finalized"
		transcript.VersionID++
		_ = db.Save(&transcript)
	}

	// Delete old simulated segments
	_ = db.Where("transcript_id = ?", transcript.ID).Delete(&entity.TranscriptSegment{})

	// Insert 226 real verbatim segments
	var newSegments []entity.TranscriptSegment
	for i, diag := range usecase.RealCourtVerbatimSegments {
		newSegments = append(newSegments, entity.TranscriptSegment{
			ID:               uuid.New(),
			TranscriptID:     transcript.ID,
			SpeakerTag:       diag.Speaker,
			StartTimeMs:      diag.StartMs,
			EndTimeMs:        diag.EndMs,
			TextContent:      diag.Text,
			SequenceOrder:    i + 1,
			IsEditedManually: false,
		})
	}

	if err := db.CreateInBatches(newSegments, 50).Error; err != nil {
		log.Fatalf("Insert real segments failed: %v", err)
	}

	log.Println("5. Updating Court Report (Berita Acara Sidang)...")
	realSummary := `Pada hari ini telah diselenggarakan persidangan perkara pidana Nomor Register 51/Pid.B/2026/PN.Tlm dengan agenda Pemeriksaan Keterangan Terdakwa atas nama Alim.

Berdasarkan pemeriksaan di hadapan Majelis Hakim dan Penasihat Hukum, diperoleh fakta-fakta persidangan sebagai berikut:
1. Terdakwa mengakui telah mengambil 1 (satu) unit sepeda motor Yamaha X-Ride yang kuncinya tergantung di dekat area perkebunan Desa Momuo, Kecamatan Paguyaman, Kabupaten Boalemo, Gorontalo saat sedang melarikan diri dari kejaran aparat kepolisian.
2. Sepeda motor tersebut kemudian dibawa lari oleh Terdakwa menuju daerah Dumoga dan Lolak (Kabupaten Bolaang Mongondow, Sulawesi Utara), lalu digadaikan di sebuah tempat bengkel motor seharga Rp 1.000.000,- (satu juta rupiah).
3. Uang hasil gadai motor tersebut digunakan Terdakwa untuk ongkos perjalanan menaiki mobil ke Pelabuhan Manado, kemudian berlayar menuju Ternate, Maluku Utara, hingga uang tersebut habis terpakai.
4. Terdakwa sempat bekerja di Maluku selama beberapa hari sebagai juru masak (koki) dengan penghasilan Rp 5.000.000,- per bulan.
5. Setelah dihubungi oleh pihak keluarga/ibu angkat melalui telepon agar bersikap kooperatif dan menyerahkan diri, Terdakwa berniat kembali ke Gorontalo namun kemudian berhasil diamankan oleh petugas kepolisian di Pelabuhan Semut Ternate, selanjutnya dijemput oleh jajaran Satreskrim Polres Boalemo.
6. Terdakwa mengakui pernah dijatuhi hukuman pidana sebelumnya (residivis tahun 2020) atas perkara pencurian.
7. Di persidangan, Terdakwa menyatakan penyesalan mendalam dan membenarkan barang bukti (BB) berupa 1 (satu) unit sepeda motor Yamaha X-Ride warna hijau.`

	realObjections := `1. Penasihat Hukum Terdakwa mengonfirmasi perihal perkara sebelumnya terkait uang keluarga yang telah dikembalikan seutuhnya dan pihak keluarga telah memberikan permaafan.
2. Penasihat Hukum menegaskan bahwa Terdakwa bertindak secara spontan tanpa perencanaan awal (dolus repentinus) saat melihat kunci tergantung pada sepeda motor karena kondisi panik dalam pelarian.
3. Penasihat Hukum memohon kepada Majelis Hakim agar sikap kooperatif Terdakwa, pengakuan terus terang, dan penyesalan yang sungguh-sungguh dicatat dalam Berita Acara sebagai pertimbangan hal-hal yang meringankan (mitigating factors).`

	var report entity.CourtReport
	if err := db.Where("court_session_id = ?", session.ID).First(&report).Error; err == nil {
		report.ReportTitle = "Berita Acara Sidang (Pemeriksaan Keterangan Terdakwa)"
		report.DocumentNumber = "BAS/51.Pid.B/2026/PN.Tlm"
		report.CaseSummary = realSummary
		report.ObjectionsNotes = realObjections
		_ = db.Save(&report)
	}

	log.Printf("SUCCESS! 226 Real verbatim segments & Case data updated for %s", courtCase.CaseNumber)
}

