package his

import "context"

type hisMockClient struct {
	// hospitalCode → id → patient
	data map[string]map[string]*HISPatient
}

func newHISMockClient() *hisMockClient {
	return &hisMockClient{
		data: map[string]map[string]*HISPatient{
			"hospital-a": {
				"1234567890123": {
					PatientHN:   "HNA001",
					NationalID:  "1234567890123",
					FirstNameTH: "สมชาย",
					LastNameTH:  "ใจดี",
					FirstNameEN: "Somchai",
					LastNameEN:  "Jaidee",
					DateOfBirth: "1990-05-20",
					Gender:      "M",
					PhoneNumber: "0812345678",
					Email:       "somchai@example.com",
				},
				"9876543210987": {
					PatientHN:   "HNA002",
					NationalID:  "9876543210987",
					FirstNameTH: "สมหญิง",
					LastNameTH:  "รักดี",
					FirstNameEN: "Somying",
					LastNameEN:  "Rakdee",
					DateOfBirth: "1985-03-15",
					Gender:      "F",
				},
				"AA1234567": {
					PatientHN:   "HNA003",
					PassportID:  "AA1234567",
					FirstNameEN: "John",
					LastNameEN:  "Smith",
					DateOfBirth: "1978-11-30",
					Gender:      "M",
				},
			},
			"hospital-b": {
				"1111111111111": {
					PatientHN:   "HNB001",
					NationalID:  "1111111111111",
					FirstNameTH: "วิชัย",
					LastNameTH:  "แสงทอง",
					FirstNameEN: "Wichai",
					LastNameEN:  "Sangthong",
					DateOfBirth: "1995-07-22",
					Gender:      "M",
				},
				"BB9876543": {
					PatientHN:   "HNB002",
					PassportID:  "BB9876543",
					FirstNameEN: "Emily",
					LastNameEN:  "Johnson",
					DateOfBirth: "2000-01-10",
					Gender:      "F",
				},
			},
		},
	}
}

func (m *hisMockClient) SearchPatient(_ context.Context, hospitalCode, id string) (*HISPatient, error) {
	hospitalData, ok := m.data[hospitalCode]
	if !ok {
		return nil, ErrHISNotConfigured
	}
	patient, ok := hospitalData[id]
	if !ok {
		return nil, ErrPatientNotFound
	}
	return patient, nil
}
