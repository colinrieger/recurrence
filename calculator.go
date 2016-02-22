package recurrence

import (
	"time"
)

const (
	day = 24 * time.Hour
)

func (p Recurrence) GetNextDate(d time.Time) time.Time {
	if p.Interval == 0 {
		p.Interval = 1
	}
	if p.Location == nil {
		p.Location = time.UTC
	}
	if p.End.After(p.Start) && !p.End.After(d) {
		return time.Time{}
	}

	switch p.Frequence {
	case NotRepeating:
		if p.Start.After(d) {
			return p.Start.In(p.Location)
		}
	case Daily:
		return p.ndDaily(d)
	case Weekly:
		return p.ndWeekly(d)
	case MonthlyXth:
		return p.ndMonthlyX(d)
	case Monthly:
		return p.ndMonthly(d)
	}
	return time.Time{}
}

func (p Recurrence) dateOf(t time.Time) time.Time {
	y, m, d := t.In(p.Location).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, p.Location)
}

func (p Recurrence) ndDaily(d time.Time) time.Time {
	start := p.Start.In(p.Location)
	end := p.End.In(p.Location)

	if d.Before(start) {
		return start
	}

	startDate := p.dateOf(start)
	timeOfDay := start.Sub(startDate)
	d = d.In(p.Location)

	dateOfD := p.dateOf(d)

	daysBetween := int(dateOfD.Sub(startDate) / day)
	freq := int(p.Interval)
	daysToAdd := (freq - (daysBetween % freq)) % freq

	res := dateOfD.Add(time.Duration(daysToAdd)*day + timeOfDay)

	if !res.After(d) {
		res = res.Add(time.Duration(p.Interval) * day)
	}
	if end.After(start) && res.After(end) {
		return time.Time{}
	}
	return res
}

func (p Recurrence) ndWeekly(d time.Time) time.Time {
	start := p.Start.In(p.Location)
	end := p.End.In(p.Location)
	d = d.In(p.Location)

	startDate := p.dateOf(start)
	timeOfDay := start.Sub(startDate)

	startOfWeek, _ := IntToWeeklyPattern(p.Pattern)
	days := p.Pattern & 255

	weekStart := startDate.Add(time.Duration(-(7+int(start.Weekday()-startOfWeek))%7) * day)
	if d.Before(weekStart) {
		d = weekStart
	}
	cycleLength := time.Duration(p.Interval*7) * day

	// Skip already passed cycles.
	weekStart = weekStart.Add(time.Duration(int(d.Sub(weekStart)/cycleLength)) * cycleLength)
	dayOfD := p.dateOf(d)

outerLoop:
	for ws := weekStart; end.Before(start) || !end.Before(ws); ws = ws.Add(cycleLength) {
		for i := 0; i < 7; i++ {
			dat := ws.Add(time.Duration(i) * day)
			if dat.Before(dayOfD) || dat.Before(startDate) {
				continue
			}

			wd := int(1 << uint(dat.Weekday()))
			if (days & wd) != wd {
				continue
			}
			dat = dat.Add(timeOfDay)
			if dat.Before(d) {
				continue
			}
			if end.After(start) && dat.After(end) {
				break outerLoop
			}
			return dat
		}
	}
	return time.Time{}
}

func (p Recurrence) ndMonthlyX(d time.Time) time.Time {
	start := p.Start.In(p.Location)
	end := p.End.In(p.Location)
	d = d.In(p.Location)
	if d.Before(start) {
		return start
	}

	dy := d.Year()
	dm := int(d.Month())

	sy := start.Year()
	sm := int(start.Month())

	interval := int(p.Interval)

	monthsBetween := ((dy - sy) * 12) + (dm - sm)
	monthsToAdd := (monthsBetween / interval) * interval
	extraIntervals := 0

	for dat := start.AddDate(0, monthsToAdd, 0); end.Before(start) || !end.Before(dat); dat = start.AddDate(0, monthsToAdd+(extraIntervals*interval), 0) {
		extraIntervals += 1
		if dat.Day() != start.Day() {
			continue
		}
		if !dat.After(d) {
			continue
		}
		return dat
	}

	return time.Time{}
}

func (p Recurrence) ndMonthly(d time.Time) time.Time {
	occ, wd := IntToMonthlyPattern(p.Pattern)

	start := p.Start.In(p.Location)
	startDate := p.dateOf(start)
	timeOfDay := start.Sub(startDate)
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, p.Location)
	end := p.End.In(p.Location)
	dStart := d.In(p.Location)
	if d.Before(start) {
		dStart = start
	}

	dy := dStart.Year()
	dm := int(dStart.Month())
	sy := start.Year()
	sm := int(start.Month())

	interval := int(p.Interval)

	monthsBetween := ((dy - sy) * 12) + (dm - sm)
	monthsToAdd := (monthsBetween / interval) * interval

	dat := start.AddDate(0, monthsToAdd, 0)
	if dat.Before(p.Start) {
		dat = dat.AddDate(0, interval, 0)
	}
	dStart = dat

	for dat.Weekday() != wd {
		dat = dat.Add(1 * day)
	}

	for !dat.After(d) {
		for i := First; i < occ; i++ {
			next := dat.Add(7 * day)
			if next.Month() != dStart.Month() {
				if occ != Last {
					dStart = dStart.AddDate(0, interval, 0)
					dat = dStart

					for dat.Weekday() != wd {
						dat = dat.Add(1 * day)
					}

					i = First
				} else {
					break
				}
			} else {
				dat = next
			}
		}
	}
	if !end.Before(p.Start) && end.Before(dat) {
		return time.Time{}
	}
	return dat.Add(timeOfDay)
}
