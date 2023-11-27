from sqlalchemy import ForeignKey, Column, Integer, String
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

class Base(DeclarativeBase):
    pass

class Subject(Base):
    __tablename__ = "subjects"
    name: Mapped[String] = mapped_column(primary_key=True)
    friendlyName: Mapped[String] = mapped_column()

class Course(Base):
    __tablename__ = "courses"
    crn: Mapped[String] = mapped_column(primary_key=True)
    id: Mapped[String] = mapped_column()
    name: Mapped[String] = mapped_column()
    section: Mapped[String] = mapped_column()
    dateRange: Mapped[String] = mapped_column()
    type: Mapped[String] = mapped_column()
    instructor: Mapped[String] = mapped_column()

class CourseTime(Base):
    __tablename__ = "courseTimes"
    crn: Mapped[String] = mapped_column(ForeignKey("courses.crn", ondelete="cascade"))
    days: Mapped[String] = mapped_column()
    startTime: Mapped[String] = mapped_column()
    endTime: Mapped[String] = mapped_column()
    location: Mapped[String] = mapped_column()
    ignore = Column(Integer, primary_key=True) #sqlalchemy NEEDS a primary key