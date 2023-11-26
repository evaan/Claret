from sqlalchemy import ForeignKey, Column, Integer
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

class Base(DeclarativeBase):
    pass

class Subject(Base):
    __tablename__ = "subjects"
    name: Mapped[str] = mapped_column(primary_key=True)
    friendlyName: Mapped[str] = mapped_column()

class Course(Base):
    __tablename__ = "courses"
    crn: Mapped[str] = mapped_column(primary_key=True)
    id: Mapped[str] = mapped_column()
    name: Mapped[str] = mapped_column()
    section: Mapped[str] = mapped_column()
    dateRange: Mapped[str] = mapped_column()
    type: Mapped[str] = mapped_column()
    instructor: Mapped[str] = mapped_column()

class CourseTime(Base):
    __tablename__ = "courseTimes"
    crn: Mapped[str] = mapped_column(ForeignKey("courses.crn", ondelete="cascade"))
    days: Mapped[str] = mapped_column()
    startTime: Mapped[str] = mapped_column()
    endTime: Mapped[str] = mapped_column()
    location: Mapped[str] = mapped_column()
    ignore = Column(Integer, primary_key=True) #sqlalchemy NEEDS a primary key