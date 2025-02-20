package db

import (
	"fmt"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
)

func (m *Migrations) removeOrphanedOttCaEnrollments(step *boltz.MigrationStep) {
	logtrace.LogWithFunctionName()
	var enrollmentsToDelete []string

	for cursor := m.stores.Enrollment.IterateIds(step.Ctx.Tx(), ast.BoolNodeTrue); cursor.IsValid(); cursor.Next() {
		current := cursor.Current()
		currentEnrollmentId := string(current)

		enrollment, err := m.stores.Enrollment.LoadById(step.Ctx.Tx(), currentEnrollmentId)

		if err != nil {
			step.SetError(fmt.Errorf("error iterating ids of enrollments, enrollment [%s]: %v", currentEnrollmentId, err))
			return
		}

		if enrollment.CaId != nil && *enrollment.CaId != "" {
			_, err := m.stores.Ca.LoadById(step.Ctx.Tx(), *enrollment.CaId)

			if err != nil && boltz.IsErrNotFoundErr(err) {
				enrollmentsToDelete = append(enrollmentsToDelete, currentEnrollmentId)
			}
		}
	}

	//clear caIds that are invalid via CheckIntegrity
	err := m.stores.Enrollment.CheckIntegrity(step.Ctx, true, func(err error, fixed bool) {
		if !fixed {
			pfxlog.Logger().Errorf("unfixable error during orphaned ottca enrollment integrity check: %v", err)
		}
	})
	step.SetError(err)

	for _, enrollmentId := range enrollmentsToDelete {
		pfxlog.Logger().Infof("removing invalid ottca enrollment [%s]", enrollmentId)
		if err := m.stores.Enrollment.DeleteById(step.Ctx, enrollmentId); err != nil {

			step.SetError(fmt.Errorf("could not delete enrollment [%s] with invalid CA reference: %v", enrollmentId, err))
		}
	}
}
