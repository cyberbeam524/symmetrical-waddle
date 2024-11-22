// pages/create-task.js
import React, { useState } from 'react';
import DataSourceSelector from '../components/DataSourceSelector';
import DataTransformationConfigurator from '../components/DataTransformationConfigurator';
import DeduplicationSettings from '../components/DeduplicationSettings';
import TaskScheduler from '../components/TaskScheduler';
import NotificationSettings from '../components/NotificationSettings';
import Navbar from '../components/Navbar'; // Ensure Navbar is in components
import { useRouter } from 'next/router';

const CreateTaskPage = () => {
  const [dataSource, setDataSource] = useState(null);
  const [transformationConfig, setTransformationConfig] = useState(null);
  const [deduplicationConfig, setDeduplicationConfig] = useState(null);
  const [scheduleConfig, setScheduleConfig] = useState(null);
  const [notificationConfig, setNotificationConfig] = useState(null);

  const router = useRouter();

  const handleCreateTask = () => {
    const taskData = {
      dataSource,
      transformationConfig,
      deduplicationConfig,
      scheduleConfig,
      notificationConfig,
    };

    // Replace with your actual API endpoint
    fetch('/api/tasks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(taskData),
    })
      .then(response => {
        if (response.ok) {
          alert('Task created successfully!');
          router.push('/'); // Redirect to homepage or dashboard
        } else {
          alert('Failed to create task.');
        }
      });
  };

  return (
    <div>
      <Navbar />
      <div className="container mx-auto mt-24 p-8">
        <h1 className="text-3xl font-bold mb-8">Create New Task</h1>
        {!dataSource && <DataSourceSelector onSelect={setDataSource} />}
        {dataSource && !transformationConfig && (
          <DataTransformationConfigurator
            dataSource={dataSource}
            onConfigure={setTransformationConfig}
          />
        )}
        {transformationConfig && !deduplicationConfig && (
          <DeduplicationSettings
            availableFields={transformationConfig.map(mapping => mapping.targetField)}
            onConfigure={setDeduplicationConfig}
          />
        )}
        {deduplicationConfig && !scheduleConfig && (
          <TaskScheduler onSchedule={setScheduleConfig} />
        )}
        {scheduleConfig && !notificationConfig && (
          <NotificationSettings onConfigure={setNotificationConfig} />
        )}
        {notificationConfig && (
          <button
            onClick={handleCreateTask}
            className="bg-blue-600 text-white px-6 py-3 rounded-lg mt-6 hover:bg-blue-700 transition"
          >
            Create Task
          </button>
        )}
      </div>
    </div>
  );
};

export default CreateTaskPage;
